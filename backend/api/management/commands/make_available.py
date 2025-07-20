from api.models import User
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = "Makes user available for algorithm to pick by removing evaluation and blocking status"

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
            user.has_eval = False
            user.blocked_until = None
            user.save()
            print(self.style.SUCCESS(f"User '{username}' has been activated."))
        except User.DoesNotExist:
            print(self.style.ERROR(f"User '{username}' does not exist."))
