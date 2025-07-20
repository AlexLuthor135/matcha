from api.models import User
from django.db import transaction
from intra_client.services.logging import peersphere_logger
from rest_framework.response import Response

def reset_has_eval():
    with transaction.atomic():
        User.objects.update(has_eval=False)
        peersphere_logger.log_admins("✅ has_eval value reset for all users.")
        return Response({"message": "has_eval value was reset in DB."})