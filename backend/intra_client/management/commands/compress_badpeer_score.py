from api.models import User
from intra_client.services.intra_utils import *
from django.db import transaction
from django.core.management.base import BaseCommand
from intra_client.services.get_corrector_score import get_corrector_score

class Command(BaseCommand):
    help = """Compresses the baspeer score for all users in the DB by the factor given as the argument."""

    def add_arguments(self, parser):
        parser.add_argument(
            "factor",
            type=int,
            help="Factor to use for the compression"
        )

    def handle(self, *args, **kwargs):
        factor = kwargs["factor"]
        with transaction.atomic():
            users = User.objects.all()
            for user in users:
                if user.student_login:
                    badpeer_score = user.badpeer_score
                    user.badpeer_score = round(badpeer_score / factor)
                    user.save()
                    print(f"User {user.student_login} has been updated. {badpeer_score} -> {user.badpeer_score}")
                else:
                    print(f"User {user.student_login} does not exist. Skipping.")
