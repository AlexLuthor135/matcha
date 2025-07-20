from api.models import User
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.core.management.base import BaseCommand
from django.conf import settings

class Command(BaseCommand):
    help = """Populate the database with currently active students
    For it to work properly from outside the docker environment, uncomment the BASE_DIR path in backend.settings"""

    def handle(self, *args, **kwargs):
        users_data = get_current_students(ic, settings.CAMPUS_ID)
        for user in users_data:
            login = user.get("user", {}).get("login", None)
            print(f"User: {login}")
        if input(f"Found {len(users_data)} users. Proceed? (y/n):") != "y":
            return
        for user in users_data:
            login = user.get("user", {}).get("login", None)
            student_id = user.get("user", {}).get("id", None)
            if login and student_id:
                try:
                    user, created = User.objects.get_or_create(
                        student_login=login,
                        defaults={
                            'student_id': student_id,
                            'username': login,
                            'happeer_score': 0
                        }
                    )
                    if created:
                        print(f"Created user: {login} with ID: {student_id}")
                    else:
                        print(f"User already exists: {login} with ID: {user.student_id}")
                except User.MultipleObjectsReturned:
                    print(f"Multiple users found with login: {login}")
                except Exception as e:
                    print(f"Error creating user {login}: {e}")
