from django.shortcuts import render
from rest_framework.generics import ListAPIView
from rest_framework.views import APIView
from rest_framework.response import Response
from admins.services.populate_locations import populate_locations
from admins.services.populate_friends import populate_friends
from admins.services.populate_cohort import populate_cohort
from admins.services.populate_ranks import populate_ranks
from admins.services.populate_avatars import populate_avatars
from admins.services.reset_has_eval import reset_has_eval
from admins.services.reset_eval_points import reset_eval_points
from admins.services.update_pending_evals import update_pending_evals
from rest_framework.permissions import IsAdminUser
from api.models import User
from admins.serializers import UserSerializer
from rest_framework import status

COMMAND_HANDLERS = {
    'populate_locations': populate_locations,
    'populate_friends': populate_friends,
    'populate_cohort': populate_cohort,
    'populate_ranks': populate_ranks,
    'populate_avatars': populate_avatars,
    'reset_has_eval': reset_has_eval,
    'reset_eval_points': reset_eval_points,
    'update_pending_evals': update_pending_evals,
}

class AdminUserListView(ListAPIView):
    queryset = User.objects.filter(has_location=True)
    serializer_class = UserSerializer
    permission_classes = [IsAdminUser]

class AdminCommandsView(APIView):
    permission_classes = [IsAdminUser]

    def post(self, request):
        command = request.data.get('command')
        if not command:
            return Response({"error": "Missing 'command' in request body."}, status=status.HTTP_400_BAD_REQUEST)
        handler = COMMAND_HANDLERS.get(command)
        if handler is None:
            return Response({"error": f"Unknown command: {command}"}, status=status.HTTP_400_BAD_REQUEST)

        return handler()
        