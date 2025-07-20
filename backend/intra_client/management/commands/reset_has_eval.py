from api.models import User
from django.core.management.base import BaseCommand
from django.db import transaction

class Command(BaseCommand):
    help = """Update the locations of students in campus
    For it to work properly from outside the docker environment, uncomment the BASE_DIR path in backend.settings"""

    def handle(self, *args, **kwargs):
        with transaction.atomic():
            User.objects.update(has_eval=False)
            print(f"has_eval value was reset in DB.")