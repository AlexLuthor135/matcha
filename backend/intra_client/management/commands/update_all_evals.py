from api.models import Evaluation, User
from django.core.management.base import BaseCommand
from django.core.management import call_command

class Command(BaseCommand):
    help = "Updates all Evaluations and sets their final_mark."

    def handle(self, *args, **options):
        for eval in Evaluation.objects.exclude(state=Evaluation.State.PENDING):
            eval.update_from_api()

        call_command('update_pending_evals')
