from rest_framework.response import Response
from intra_client.services.intra_api import ic
from intra_client.views.intra_team import IntraTeams

class IntraTeamsListView(IntraTeams):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.user = None
        
    def get(self, request):
        try:
            self.user = request.user
            extracted_teams = self.get_teams()
            return Response({
                "teams": extracted_teams,
                "user_info": self.get_user_info(),
            }, status=200)
        except Exception as e:
            return Response({"error": str(e)}, status=500)
        
    def get_teams(self):
        student_id = self.user.student_id
        teams = ic.pages_threaded(f'/users/{student_id}/teams', params={
            "filter[status]": "waiting_for_correction",
        })
        extracted_teams = []
        for team in teams:
            is_leader = self.get_leader(team)
            if is_leader:
                project = ic.pages_threaded(f'/projects/{team.get("project_id")}')
                scale_teams = self.get_non_truant_scale_teams(team)
                correction_number = (
                    ic.pages_threaded(f'/scales/{scale_teams[0]["scale_id"]}').get("correction_number")
                    if scale_teams else None
                )
                extracted_teams.append({
                    "id": team.get("id"),
                    "name": team.get("name"),
                    "project_name": project.get("name"),
                    "scale_teams": str(len(scale_teams)),
                    "correction_number": correction_number,
                })
        return extracted_teams
    
    def get_user_info(self):
        try:
            user_info = {
                "weekly_evals": self.user.weekly_evals,
                "blocked_until": self.user.blocked_until.isoformat() if self.user.blocked_until else None,
                "block_pass": self.user.block_pass,
                "block_pass_used_at": self.user.block_pass_used_at.isoformat() if self.user.block_pass_used_at else None,
                "available_pass": self.user.available_pass,
            }
            return user_info
        except Exception as e:
            return Response({"error": str(e)}, status=500)

