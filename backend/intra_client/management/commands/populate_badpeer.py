from api.models import User
from intra_client.services.intra_utils import *
from django.db import transaction
from django.core.management.base import BaseCommand
from intra_client.services.get_corrector_score import get_corrector_score

class Command(BaseCommand):
    help = """Populate the database with badpeer score - the amount of evals student performed against the ones they got.
    For it to work properly from outside the docker environment, uncomment the BASE_DIR path in backend.settings"""

    def handle(self, *args, **kwargs):
        with transaction.atomic():
            User.objects.update(badpeer_score=0)
            users = User.objects.all()
            for user in users:
                if user.student_login:
                    happeer_score = get_corrector_score(user.student_login, 'as_corrector')
                    badpeer_score = get_corrector_score(user.student_login, 'as_corrected')
                    user.badpeer_score = happeer_score - badpeer_score
                    user.save()
                    print(f"User {user.student_login} has been updated. Badpeer score: {badpeer_score}")
                else:
                    print(f"User {user.student_login} does not exist. Skipping.")