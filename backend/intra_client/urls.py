from django.urls import path, include
from intra_client.views.intra_teams.assign_view import IntraTeamsAssignView
from intra_client.views.intra_teams.list_view import IntraTeamsListView
from intra_client.views.intra_webhook import IntraWebhook

urlpatterns = [
    path('teams/', IntraTeamsListView.as_view(), name='intra_teams'),
    path('teams/assign/', IntraTeamsAssignView.as_view(), name='intra_teams'),
    path('webhook/', IntraWebhook.as_view(), name='intra_webhook'),
]
