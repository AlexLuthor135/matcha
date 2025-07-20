from api.models import User
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.core.management.base import BaseCommand
from django.conf import settings
from django.db import transaction

class Command(BaseCommand):
    help = """Update the locations of students in campus
    For it to work properly from outside the docker environment, uncomment the BASE_DIR path in backend.settings"""

    def handle(self, *args, **kwargs):
        users_data = get_users_with_active_location(ic, settings.CAMPUS_ID)
        if not users_data:
            print("No users found.")
            return
        if input(f"Found {len(users_data)} users. Proceed? (y/n):") != "y":
            return
        with transaction.atomic():
            User.objects.update(has_location=False)
            updated_count = 0
            for user_data in users_data:
                login = user_data.get("user").get("login")
                if not login:
                    continue
                try:
                    user = User.objects.get(student_login=login)
                    user.has_location = True
                    user.save()
                    updated_count += 1
                    print(f"User {login} has been updated.")
                except User.DoesNotExist:
                    print(f"User {login} does not exist. Skipping.")
            print(f"Updated {updated_count} users.")