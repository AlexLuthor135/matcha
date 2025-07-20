from api.models import User
from intra_client.services.intra_utils import *
from django.db import transaction
from django.core.management.base import BaseCommand
from intra_client.services.intra_api import ic

AMOUNT_TO_REMOVE = "1000"
GRANTED_REASON = "PeerSphere reset."

class Command(BaseCommand):
    help = """Reset the eval points to the provided number"""
    
    def handle(self, *args, **kwargs):
        with transaction.atomic():
            users = User.objects.all()
            for user in users:
                if user.student_id:
                    payload = {
                        "id": user.student_id, 
                        "reason": GRANTED_REASON,
                        "amount": AMOUNT_TO_REMOVE
                    }
                    try:
                        response = ic.delete(f'users/{user.student_id}/correction_points/remove', json=payload)
                        print(f"{user.student_login}: 1000 points removed.")
                    except Exception as e:
                        print(f"Failed to remove the eval points: {response.status_code}")
                        print (f"Error: {e}")
                else:
                    print(f"User {user.student_login} does not exist. Skipping.")