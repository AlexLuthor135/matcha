from api.models import User
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = "Force-activate a user by setting has_location=True"

    def add_arguments(self, parser):
        parser.add_argument(
            "username",
            type=str,
            help="Username of the user to activate"
        )

    def handle(self, *args, **options):
        username = options["username"]
        try:
            user = User.objects.get(username=username)
            if user.has_location:
                print(self.style.WARNING(f"User '{username}' is already active."))
            else:
                user.has_location = True
                user.save()
                print(self.style.SUCCESS(f"User '{username}' has been activated."))
        except User.DoesNotExist:
            print(self.style.ERROR(f"User '{username}' does not exist."))
