from api.models import User
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.db import transaction
from django.core.management.base import BaseCommand
from django.conf import settings

class Command(BaseCommand):
    help = """Populate the database with quest users
    For it to work properly from outside the docker environment, uncomment the BASE_DIR path in backend.settings"""


    def process_users(self, users, user_logins, unique_logins):
        for user in users:
            login = user["login"]
            if login not in unique_logins:
                user_logins.append(login)
                unique_logins.add(login)
    
    def handle(self, *args, **kwargs):
        unique_logins = set()
        ranks_and_logins = {
        7: [],
        6: [],
        5: [],
        4: [],
        3: [],
        2: [],
        1: [],
        0: []
        }

        quests_and_rank_ids = [
        (7, get_quest_users(ic, f"{settings.CAMPUS_ID}", 37, None, True)),
        (6, get_quest_users(ic, f"{settings.CAMPUS_ID}", 49)),
        (5, get_quest_users(ic, f"{settings.CAMPUS_ID}", 48)),
        (4, get_quest_users(ic, f"{settings.CAMPUS_ID}", 47)),
        (3, get_quest_users(ic, f"{settings.CAMPUS_ID}", 46)),
        (2, get_quest_users(ic, f"{settings.CAMPUS_ID}", 45)),
        (1, get_quest_users(ic, f"{settings.CAMPUS_ID}", 44)),
        (0, get_quest_users(ic, f"{settings.CAMPUS_ID}", 37)),
        ]

        for rank_number, quest_users in quests_and_rank_ids:
            if quest_users:
                self.process_users(quest_users, ranks_and_logins[rank_number], unique_logins)

        with transaction.atomic():
            User.objects.update(quest_rank=None)

            for rank_number, user_logins in sorted(ranks_and_logins.items(), reverse=True):
                for login in user_logins:
                    try:
                        user = User.objects.get(username=login)
                        user.quest_rank = rank_number
                        user.save()
                    except User.DoesNotExist:
                        print(f"User {login} does not exist. Skipping.")
        for rank_number in sorted(ranks_and_logins.keys(), reverse=True):
            user_count = User.objects.filter(quest_rank=rank_number).count()
            print(f"Rank {rank_number}: {user_count} users.")
    
