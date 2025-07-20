from api.models import Evaluation
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = """Creates test data in Evaluations model"""

    def handle(self, *args, **kwargs):
        Evaluation.objects.create(
        id="1234",
        team=["Alice", "Bob", "Carlos"],
        evaluator="Diana",
        project="AI Chatbot",
        state=Evaluation.State.PENDING
)
