from django.conf import settings
from slack.services.slack_utils import *
from intra_client.services.scale_team_class import ScaleTeam
from intra_client.services.logging import peersphere_logger

def get_scale_team_info_format_string(scale_team:ScaleTeam) -> str:
    evaluatee_str = ' '.join([f"`<https://profile.intra.42.fr/users/{evaluatee}|{evaluatee}>`" for evaluatee in scale_team.users])
    scale_team_str = '\n'.join([
        f">*Group Members:* {evaluatee_str}",
        f">*Evaluator:* `<https://profile.intra.42.fr/users/{scale_team.evaluator}|{scale_team.evaluator}>`",
        f">*Project:* <https://projects.intra.42.fr/projects/{scale_team.project_slug}/projects_users/{scale_team.project_user_id}|{scale_team.project_slug}>",
    ])
    return scale_team_str

def send_messege_to_user(user:str, msg:str):
    user_email = f"{user}{settings.EMAIL_STRING}"
    user_id = get_slack_user_id_by_email(user_email, settings.SLACK_BORN2CODE_TOKEN)
    peersphere_logger.log_slack(f"Get Id: {user} -> {user_email} -> {user_id}")
    if not user_id:
        raise ValueError(f"Failed to get Slack user ID for email: {user_email}")
    slack_post_message("", user_id, [[
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": msg
            }
        }
    ]], settings.SLACK_BORN2CODE_TOKEN)
