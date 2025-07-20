from api.models import User
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = "Setting has_location=False for all users"

    def handle(self, *args, **options):
        try:
            users = User.objects.all()
            for user in users:
                user.has_location = False
                user.save()
        except User.DoesNotExist:
            print(self.style.ERROR(f"User '{user}' does not exist."))
