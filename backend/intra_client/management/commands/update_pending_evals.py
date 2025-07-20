from api.models import Evaluation, User
from django.core.management.base import BaseCommand
from django.utils import timezone
from datetime import timedelta

class Command(BaseCommand):
    help = "Checks all pending evals for missed webhooks. Should be run regularly."

    def handle(self, *args, **options):
        evals = Evaluation.objects.filter(state=Evaluation.State.PENDING)
        print(f"Checking {evals.count()} Evals.")
        for eval in evals:
            if eval.update_from_api():
                print(f"Updated {eval.id} from {eval.time_created} -> {eval.state}")

        users = User.objects.filter(has_eval=True)
        print(f"Checking {users.count()} users.")
        for user in users:
            if Evaluation.objects.filter(evaluator=user.username, state=Evaluation.State.PENDING).count() == 0:
                user.has_eval = False
                user.save()
                print(f"Updated user {user.username}")
