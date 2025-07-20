from api.models import User
from django.core.management.base import BaseCommand
from django.utils import timezone
from django.utils.dateparse import parse_datetime

class Command(BaseCommand):
    help = "Block list of users until the given time."

    def handle(self, *args, **options):

        blocked_until_time = parse_datetime("2025-06-02 23:00:00+00")
        logins = [
            "feberman",
        ]
        for login in logins:
            try:
                user = User.objects.get(username=login)
                user.blocked_until = blocked_until_time
                user.save()
                print(self.style.SUCCESS(f"User '{login}' has been blocked."))
            except User.DoesNotExist:
                print(self.style.ERROR(f"User '{login}' does not exist."))
