from api.models import User
from intra_client.services.intra_api import ic
from rest_framework.response import Response
from intra_client.services.logging import peersphere_logger
from concurrent.futures import ThreadPoolExecutor
from django.db import close_old_connections

AMOUNT_TO_REMOVE = "1000"
GRANTED_REASON = "PeerSphere reset."

def process_eval_points(user):
    close_old_connections()
    if user.student_id:
        payload = {
            "id": user.student_id, 
            "reason": GRANTED_REASON,
            "amount": AMOUNT_TO_REMOVE
        }
        try:
            response = ic.delete(f'users/{user.student_id}/correction_points/remove', json=payload)
            peersphere_logger.log_admins(f"✅ {user.student_login}: 1000 points removed.")
            return 1
        except Exception as e:
            peersphere_logger.log_admins(f"❌ Error removing eval points for {user.student_login}: {e}")
            return 0
    else:
        print(f"User {user.student_login} does not exist. Skipping.")
        return 0

def reset_eval_points():
    users = list(User.objects.all())
    with ThreadPoolExecutor(max_workers=4) as executor:
        results = list(executor.map(process_eval_points, users))
    return Response({"message": "Eval points reset successfully.", "updated": sum(results)})