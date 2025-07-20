from api.models import Evaluation
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = """Wipes the complete database. USE WITH CAUTION"""

    def handle(self, *args, **kwargs):
        confirm = input("Are you sure you want to delete ALL evaluations? Type 'yes' to continue: ")
        if confirm.lower() == 'yes':
            count, _ = Evaluation.objects.all().delete()
            print(self.style.SUCCESS(f"Successfully deleted {count} Evals."))
        else:
            print(self.style.WARNING("Operation cancelled."))
