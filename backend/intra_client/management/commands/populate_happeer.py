from api.models import User
from intra_client.services.intra_utils import *
from django.db import transaction
from django.core.management.base import BaseCommand
from intra_client.services.get_corrector_score import get_corrector_score

class Command(BaseCommand):
    help = """Populate the database with happeer score - the total amount of evaluations occurred during CC.
    For it to work properly from outside the docker environment, uncomment the BASE_DIR path in backend.settings"""
    
    def handle(self, *args, **kwargs):
        with transaction.atomic():
            User.objects.update(happeer_score=0)
            users = User.objects.all()
            for user in users:
                if user.student_login:
                    happeer_score = get_corrector_score(user.student_login, 'as_corrector')
                    user.happeer_score = happeer_score
                    user.save()
                    print(f"User {user.student_login} has been updated. Happeer score: {happeer_score}")
                else:
                    print(f"User {user.student_login} does not exist. Skipping.")