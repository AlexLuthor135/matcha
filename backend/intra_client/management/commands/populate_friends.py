from api.models import User
import requests
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.db import transaction
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = """Populate the database with past evaluators
    For it to work properly from outside the docker environment, uncomment the BASE_DIR path in backend.settings"""

    def get_past_evaluators(self, student_login):
        piscine_project_id = [1255, 1256, 1305, 1257, 1258, 1259, 1260, 1261, 1262, 1263, 1270, 1264, 1265, 1266, 1267, 1268, 1271, 2315, 1308, 1310, 1309, 2081, 2079,]
        correctors = set()
        try:
            scale_teams = ic.pages_threaded(f"users/{student_login}/scale_teams/as_corrected")
            for scale_team in scale_teams:
                if scale_team.get('truant'):
                    continue
                corrector = scale_team.get("corrector")
                if scale_team.get("team", {}).get("project_id") not in piscine_project_id and corrector and corrector.get("login"):
                    correctors.add(corrector.get("login"))
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                print(f"User {student_login} not found.")
            else:
                print(f"Error: {e}")
        except Exception as e:
            print(f"Error: {e}")
        return list(correctors)

    def handle(self, *args, **kwargs):
        with transaction.atomic():
            User.objects.update(friends=[])
            users = User.objects.all()
            for user in users:
                if user.student_login:
                    correctors = self.get_past_evaluators(user.student_login)
                    user.friends = correctors
                    user.save()
                    print(f"User {user.student_login} has been updated.")
                else:
                    print(f"User {user.student_login} does not exist. Skipping.")