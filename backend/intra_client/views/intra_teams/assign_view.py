from intra_client.views.intra_team import IntraTeams
from rest_framework.response import Response
from rest_framework.permissions import IsAuthenticated
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from api.models import User
from datetime import timedelta
from django.utils import timezone
from intra_client.models.rank_node import RankNode
from intra_client.services.night_eval import check_night_eval
from intra_client.services.logging import peersphere_logger

PEER_VIDEO_PROJECTS = [1643, 1649, 1655, 1661, 1667,]

class IntraTeamsAssignView(IntraTeams):
    permission_classes = [IsAuthenticated]
    
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.nodes = self.build_nodes()
        self.team_id = None
        self.team = None
        self.team_logins = []
        self.scale_teams = []
        self.current_rank = None
        self.correctors_login = []
        self.correctors_circle = {}
        self.past_evaluators = []
        self.active_students = []
        self.new_evaluators = {}
        self.project_id = None

    def initialize_self(self, request):
        self.team_id = request.data.get('team_id')
        self.team = self.get_team(self.team_id)
        self.scale_teams = self.get_non_truant_scale_teams(self.team)
        self.correctors_login = self.get_correctors_login(self.scale_teams)
        self.team_logins = self.get_team_logins(self.team)
        self.current_rank = User.objects.get_users_rank(self.get_leader(self.team))
        self.past_evaluators = self.get_past_evaluators()
        self.active_students = User.objects.get_active_students()
        self.correctors_circle = self.set_corrector_option()
        self.new_evaluators = self.find_evaluators()
        self.project_id = self.get_project_id(self.team)
    
    def post(self, request):
        self.initialize_self(request)
        if self.project_id not in PEER_VIDEO_PROJECTS and not request.user.has_location:
            return Response({"error": "You have to be on campus to book an evaluation"}, status=400)
        return (self.create_scale_team())
    
    def create_scale_team(self):
        all_logins = [self.new_evaluators.get(circle) for circle in ("lower", "current", "higher") if self.new_evaluators.get(circle)]
        users = [User.objects.get_user_by_login(login) for login in all_logins]
        if not users:
            peersphere_logger.log_usage(f"❌ {self.team_id} {self.team_logins}")
            return Response({"error": "No suitable evaluators found in school"}, status=400)
        best_user = min((user for user in users if user), key=lambda u: u.happeer_score, default=None)
        if not best_user:
            peersphere_logger.log_usage(f"❌ {self.team_id} {self.team_logins}")
            return Response({"error": "No suitable evaluators found in school"}, status=400)
        student_id = best_user.student_id
        payload = self.get_payload(student_id)
        headers = {
        'Content-Type': 'application/json',
        }
        response = ic.post('scale_teams/multiple_create', json=payload, headers=headers)
        best_user.blocked_until = timezone.now() + timedelta(hours=24)
        if best_user.available_pass:
            best_user.available_pass = False
        best_user.save()
        peersphere_logger.log_usage(f"✅ {self.team_id} {self.team_logins} -> {best_user} ({best_user.happeer_score})")
        return Response(response.json())

    def get_payload(self, student_id):
        now = timezone.now()
        if self.project_id not in PEER_VIDEO_PROJECTS:
            now -= timedelta(minutes=15)
            minute = (now.minute // 15) * 15
        else:
            minute = (now.minute // 15 + 1) * 15
            if minute == 60:
                now += timedelta(hours=1)
                minute = 0
        begin_at = now.replace(minute=minute, second=0, microsecond=0).isoformat() + 'Z'
        payload = {
            "scale_teams": [{
                "team_id": self.team_id,
                "begin_at": begin_at,
                "user_id": student_id,
            }]
        }
        return payload

    def set_corrector_option(self):
        """Sets the values of already occured evaluations in the team or a pending one"""
        correctors_circle = {
            "lower": False,
            "current": False,
            "higher": False
        }
        quest_ids = [rank for login in self.correctors_login if (rank := User.objects.get_users_rank(login)) is not None]
        for quest_id in quest_ids:
            if quest_id < self.current_rank:
                correctors_circle["lower"] = True
            elif quest_id == self.current_rank:
                correctors_circle["current"] = True
            else:
                correctors_circle["higher"] = True
        return correctors_circle

    def get_past_evaluators(self):
        """Get the list of past evaluators for the team members"""
        correctors = set()
        for team_member in self.team_logins:
            friends = User.objects.get_friends(team_member)
            correctors.update(friends)
        return correctors

    def process_priority(self, match_logins, group):
        """Adds student logins to the match list that are active and not having an evaluation"""
        group_logins = [user.student_login for user in group]
        for active_student in self.active_students:
            if User.objects.get_user_eval_status(active_student):
                continue
            if active_student in group_logins:
                match_logins.append(active_student)

    def find_evaluator(self, group: RankNode):
        def try_find_in_group(current_group, exclude_friends):
            """Try to find a candidate in the given group with the given friend exclusion setting."""
            match_logins = []
            self.process_priority(match_logins, current_group.users)
            filtered = [
                login for login in match_logins
                if login not in self.team_logins
                and login not in self.correctors_login
                and (not exclude_friends or login not in self.past_evaluators)
            ]
            if filtered:
                users = [User.objects.get_user_by_login(login) for login in filtered]
                return min((user for user in users if user), key=lambda u: u.happeer_score, default=None)
            return None
        exclude_order = [not check_night_eval(), False]
        for exclude_friends in exclude_order:
            current = group
            while current:
                candidate = try_find_in_group(current, exclude_friends=exclude_friends)
                if candidate:
                    return candidate
                current = current.next
        return None

    def find_evaluators(self):
        def safe_get_node(offset):
            node = self.get_node(self.current_rank + offset)
            return node or self.get_node(self.current_rank)
        lower_circle_evaluator = (self.find_evaluator(safe_get_node(-1)) if (not self.correctors_circle.get("lower", True) or self.project_id in PEER_VIDEO_PROJECTS) else None)
        current_circle_evaluator = (self.find_evaluator(safe_get_node(0)) if not self.correctors_circle.get("current", True) else None)
        higher_circle_evaluator = (self.find_evaluator(safe_get_node(1)) if not self.correctors_circle.get("higher", True) else None)
        return {"lower": lower_circle_evaluator, "current": current_circle_evaluator, "higher": higher_circle_evaluator,}
