from api.models import User
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.conf import settings
from intra_client.services.get_corrector_score import get_corrector_score
from intra_client.services.intra_utils import get_users_quest_rank
from admins.services.populate_friends import get_past_evaluators
from admins.services.populate_avatars import get_avatar
from rest_framework.response import Response
from concurrent.futures import ThreadPoolExecutor
from intra_client.services.logging import peersphere_logger
from django.db import close_old_connections

def process_user(user):
    close_old_connections()
    login = user.get("user", {}).get("login", None)
    student_id = user.get("user", {}).get("id", None)
    if login and student_id:
        try:
            user_obj, created = User.objects.get_or_create(
                student_login=login,
                defaults={
                    'student_id': student_id,
                    'username': login,
                }
            )
            if created:
                user_obj.happeer_score = get_corrector_score(login, 'as_corrector')
                user_obj.badpeer_score = user_obj.happeer_score - get_corrector_score(login, 'as_corrected')
                user_obj.friends = get_past_evaluators(login)
                user_obj.has_location = False
                user_obj.avatar = get_avatar(login)
                user_obj.quest_rank = get_users_quest_rank(ic, student_id)
                user_obj.save()
                peersphere_logger.log_admins(f"✅ User {login} created successfully.")
                return (1, 0, 0)
            else:
                peersphere_logger.log_admins(f"⚠️ User {login} already exists. Skipping.")
                return (0, 1, 0)

        except User.MultipleObjectsReturned:
            peersphere_logger.log_admins(f"❌ Multiple users found for {login}. Skipping.")
            return (0, 0, 1)
        except Exception as e:
            peersphere_logger.log_admins(f"❌ Error creating user {login}: {e}")
            return (0, 0, 1)
    return (0, 0, 0)


def populate_cohort():
    users_data = get_current_students(ic, settings.CAMPUS_ID)
    created_count = 0
    skipped_count = 0
    errors = 0
    with ThreadPoolExecutor(max_workers=4) as executor:
        results = list(executor.map(process_user, users_data))
        for created, skipped, error in results:
            created_count += created
            skipped_count += skipped
            errors += error
    return Response({
        "created_count": created_count,
        "skipped_count": skipped_count,
        "errors": errors
    })