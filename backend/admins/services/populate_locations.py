from api.models import User
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.conf import settings
from rest_framework.response import Response
from intra_client.services.logging import peersphere_logger
from concurrent.futures import ThreadPoolExecutor
from django.db import close_old_connections

def process_locations(user):
    close_old_connections()
    login = user.get("user").get("login")
    if not login:
        return 0
    try:
        user = User.objects.get(student_login=login)
        user.has_location = True
        user.save()
        peersphere_logger.log_admins(f"✅ Location updated for {login}")
        return 1
    except User.DoesNotExist:
        peersphere_logger.log_admins(f"❌ User {login} does not exist. Skipping.")
        return 0


def populate_locations():
    users_data = get_users_with_active_location(ic, settings.CAMPUS_ID)
    if not users_data:
        peersphere_logger.log_admins("❌ No active users found.")
        return Response({"message": f"No active users found."})
    User.objects.update(has_location=False)
    with ThreadPoolExecutor(max_workers=4) as executor:
        results = sum(executor.map(process_locations, users_data))
    peersphere_logger.log_admins(f"✅ Locations updated for {results} users.")
    return Response({"message": f"Locations updated. {results} users active."})
    