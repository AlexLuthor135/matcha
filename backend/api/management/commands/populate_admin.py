from api.models import User
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = "Create or update an admin user with specific login and student ID"

    def handle(self, *args, **kwargs):
        users_to_create = [
            {"login": "santiago", "student_id": 229790},
            # {"login": "alexandros", "student_id": 212170},
            # Add more here as needed
        ]

        for user_data in users_to_create:
            login = user_data["login"]
            student_id = user_data["student_id"]
            try:
                user, created = User.objects.update_or_create(
                    student_id=student_id,
                    defaults={
                        "student_id": student_id,
                        "username": login,
                        "happeer_score": 0,
                        "is_staff": True,
                        "is_superuser": True,
                    }
                )
                if created:
                    self.stdout.write(self.style.SUCCESS(f"Created user: {login}"))
                else:
                    self.stdout.write(self.style.NOTICE(f"Updated existing user: {login}"))
            except Exception as e:
                self.stdout.write(self.style.ERROR(f"Error creating user {login}: {e}"))
