from datetime import timedelta
from django.utils import timezone
from rest_framework.permissions import IsAuthenticated
from rest_framework.views import APIView
from rest_framework.response import Response
from api.models import UserAction
from intra_client.services.logging import peersphere_logger

class ResetBlockedView(APIView):
    permission_classes = [IsAuthenticated]

    def post(self, request):
        user = request.user
        user.available_pass = True
        user.save()
        UserAction.objects.create(
            user = user.student_login,
            action = UserAction.Action.UNBLOCK
        )
        peersphere_logger.log_action(f"🙋‍♂️ Unblock used by {user.student_login}")
        return Response({
            "message": "User is available for evaluations.",
        })
    
class SetBlockedView(APIView):
    permission_classes = [IsAuthenticated]

    def post(self, request):
        user = request.user
        if not user.block_pass:
            if user.available_pass:
                user.available_pass = False
                user.save()
                return Response({
                    "error": "Block pass was used for this week, available overdrive deactivated",
                },)
            return Response({
                "error": "Block pass was used for this week.",
            }, status=400)
        user.blocked_until = timezone.now() + timedelta(hours=24)
        user.block_pass_used_at = timezone.now()
        user.available_pass = False
        user.save()
        UserAction.objects.create(
            user = user.student_login,
            action = UserAction.Action.BLOCK
        )
        peersphere_logger.log_action(f"🙅‍♂️ Block used by {user.student_login}")
        return Response({
            "message": "User is unavailable for evaluations.",
        })
