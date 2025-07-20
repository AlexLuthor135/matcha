from api.models import User
import requests
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from rest_framework.response import Response
from concurrent.futures import ThreadPoolExecutor
from intra_client.services.logging import peersphere_logger
from django.db import close_old_connections

def get_past_evaluators(student_login):
        piscine_project_id = [1255, 1256, 1305, 1257, 1258, 1259, 1260, 1261, 1262, 1263, 1270, 1264, 1265, 1266, 1267, 1268, 1271, 2315, 1308, 1310, 1309, 2081, 2079,]
        correctors = set()
        try:
            scale_teams = ic.pages_threaded(f"users/{student_login}/scale_teams/as_corrected")
            for scale_team in scale_teams:
                if scale_team.get('truant'):
                    continue
                corrector = scale_team.get("corrector")
                if scale_team.get("team", {}).get("project_id") not in piscine_project_id and corrector and corrector.get("login"):
                    correctors.add(corrector.get("login"))
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                peersphere_logger.log_admins(f"❌ User {student_login} not found or has no scale teams.")
            else:
                peersphere_logger.log_admins(f"❌ HTTP error for user {student_login}: {e}")
        except Exception as e:
            peersphere_logger.log_admins(f"❌ Error fetching scale teams for user {student_login}: {e}")
        return list(correctors)

def process_friends(user):
    close_old_connections()
    if user.student_login:
        correctors = get_past_evaluators(user.student_login)
        user.friends = correctors
        user.save()
        peersphere_logger.log_admins(f"✅ Friends updated for {user.student_login}")
        return 1
    else:
        peersphere_logger.log_admins(f"❌ User {user.username} has no student login. Skipping.")
        return 0

def populate_friends():
    User.objects.update(friends=[])
    users = list(User.objects.all())
    with ThreadPoolExecutor(max_workers=4) as executor:
        results = list(executor.map(process_friends, users))
    return Response({"message": "Friends populated successfully.", "updated": sum(results)})