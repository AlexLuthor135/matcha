from api.models import Evaluation
from django.core.management.base import BaseCommand
from django.utils import timezone

class Command(BaseCommand):
    help = "Sets an evaluation to be a desired state"

    def add_arguments(self, parser):
        parser.add_argument(
            "eval_id",
            type=str,
            help="Id of the evaluation"
        )
        parser.add_argument(
            "state",
            type=str,
            help="the desired state"
        )

    def handle(self, *args, **options):
        id = int(options["eval_id"])
        state = options["state"]
        try:
            evaluation = Evaluation.objects.get(id=id)
            if state == 'pending':
                evaluation.state = Evaluation.State.PENDING
                evaluation.time_finished = None
            elif state == 'cancelled':
                evaluation.state = Evaluation.State.CANCELLED
                evaluation.time_finished = timezone.now()
            elif state == 'successful':
                evaluation.state = Evaluation.State.SUCCESSFUL
                evaluation.time_finished = timezone.now()
            else:
                print(self.style.ERROR(f"{state} not recognized"))
            evaluation.save()
            print(f"Eval {id} updated to {state}")
        except Evaluation.DoesNotExist:
            print(self.style.ERROR(f"Eval {id} does not exist."))
