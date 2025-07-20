from api.models import User
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = "Create users from a hardcoded list"

    def handle(self, *args, **kwargs):
        users_to_create = [
            {"login": "freddy", "student_id": 176923},
            {"login": "ppedro", "student_id": 191582},
            {"login": "dstud", "student_id": 113543},
            {"login": "feberman", "student_id": 147810},
            # Add more here as needed
        ]

        for user_data in users_to_create:
            login = user_data["login"]
            student_id = user_data["student_id"]

            try:
                user, created = User.objects.get_or_create(
                    student_login=login,
                    defaults={
                        "student_id": student_id,
                        "username": login,
                        "happeer_score": 0
                    }
                )
                if created:
                    self.stdout.write(self.style.SUCCESS(f"Created user: {login}"))
                else:
                    self.stdout.write(self.style.WARNING(f"User already exists: {login}"))
            except Exception as e:
                self.stdout.write(self.style.ERROR(f"Error creating user {login}: {e}"))
