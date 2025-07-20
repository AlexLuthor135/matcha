from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated
from intra_client.services.intra_utils import *
from django.contrib.auth import get_user_model
from intra_client.models.rank_node import RankNode
from intra_client.services.intra_api import ic

class IntraTeams(APIView):
    permission_classes = [IsAuthenticated]

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
    
    def get_node(self, rank):
        return self.nodes[rank] if 0 <= rank <= 7 else None

    def build_nodes(self):
        nodes = [RankNode(rank, self.get_users_by_rank(rank)) for rank in range(8)]
        for i in range(8):
            if i > 0:
                nodes[i].prev = nodes[i - 1]
            if i < 7:
                nodes[i].next = nodes[i + 1]
        return nodes

    def get_users_by_rank(self, rank):
        return get_user_model().objects.get_users_by_rank(rank)

    def get_leader(self, team):
        for user in team.get("users", []):
            if user.get("leader"):
                return user.get("login")
        return None
    
    def get_non_truant_scale_teams(self, team):
        """Extracts and returns non-truant scale teams from a full team object."""
        return [st for st in team.get("scale_teams", []) if not st.get("truant")]
    
    def get_correctors_login(self, scale_teams):
        return [
            st["corrector"]["login"]
            for st in scale_teams
            if st.get("corrector") and st["corrector"].get("login")
        ]

    def get_team_logins(self, team):
        return [user.get("login") for user in team.get("users", []) if user.get("login")]
    
    def get_team(self, team_id):
        try:
            return ic.pages_threaded(f"teams/{team_id}")
        except Exception as e:
            raise RuntimeError(f"Failed to fetch team: {str(e)}")
    
    def get_project_id(self, team):
        return team.get("project_id")


    
    



    
