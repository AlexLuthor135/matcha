from django.urls import path
from .views import AdminUserListView, AdminCommandsView

urlpatterns = [
    path('users/', AdminUserListView.as_view(), name='admin-user-list'),
    path('commands/', AdminCommandsView.as_view(), name='admin-commands'),
]
