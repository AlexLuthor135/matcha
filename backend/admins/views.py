from rest_framework.generics import ListAPIView

from rest_framework.permissions import IsAdminUser
from admins.serializers import UserSerializer

class AdminUserListView(ListAPIView):
    serializer_class = UserSerializer
    permission_classes = [IsAdminUser]