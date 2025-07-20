from api.models import User
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.db import transaction
from rest_framework.response import Response
from django.conf import settings
from intra_client.services.logging import peersphere_logger
from concurrent.futures import ThreadPoolExecutor
from django.db import close_old_connections

def process_users(users, user_logins, unique_logins):
    for user in users:
        login = user["login"]
        if login not in unique_logins:
            user_logins.append(login)
            unique_logins.add(login)

def process_ranks(args):
    close_old_connections()
    login, rank_number = args
    try:
        user = User.objects.get(username=login)
        user.quest_rank = rank_number
        user.save()
        peersphere_logger.log_admins(f"✅ Rank {rank_number} updated for {login}")
        return 1
    except User.DoesNotExist:
        peersphere_logger.log_admins(f"❌ User {login} does not exist. Skipping.")
        return 0
    
def populate_ranks():
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
            process_users(quest_users, ranks_and_logins[rank_number], unique_logins)
    with transaction.atomic():
        User.objects.update(quest_rank=None)
    update_args = []
    for rank_number, user_logins in sorted(ranks_and_logins.items(), reverse=True):
        for login in user_logins:
            update_args.append((login, rank_number))
    with ThreadPoolExecutor(max_workers=4) as executor:
        results = sum(executor.map(process_ranks, update_args))
    for rank_number in sorted(ranks_and_logins.keys(), reverse=True):
        user_count = User.objects.filter(quest_rank=rank_number).count()
        peersphere_logger.log_admins(f"✅ {user_count} users updated with rank {rank_number}")
    peersphere_logger.log_admins(f"✅ Total users updated with ranks: {results}")
    peersphere_logger.log_admins("✅ Ranks populated successfully.")
    return Response({
        "message": "Ranks populated successfully.",
        "updated": results,
        "ranks_and_logins": {rank: len(logins) for rank, logins in ranks_and_logins.items()}
    })

