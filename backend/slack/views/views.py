# from __future__ import annotations
from slack.services.slack_utils import *
from django.conf import settings
from intra_client.services.scale_team_class import ScaleTeam
from intra_client.services.logging import peersphere_logger
from .utils import get_scale_team_info_format_string, send_messege_to_user



def post_slack_scale_create(scale_team:ScaleTeam):
    post_slack_staff_scale_create(scale_team)
    post_slack_users_scale_create(scale_team)

def post_slack_scale_update_cancelled(scale_team:ScaleTeam, cancelled_eval_count:int):
    post_slack_staff_scale_update_cancelled(scale_team, cancelled_eval_count)
    post_slack_users_scale_update_cancelled(scale_team)

def post_slack_staff_scale_create(scale_team:ScaleTeam):
    header = ":scales: Scale Team Created :scales:"
    message = get_scale_team_info_format_string(scale_team)
    post_message_to_channel(header, message)

def post_slack_users_scale_create(scale_team:ScaleTeam):
    message = '\n'.join([
        f"You have been assigned a new evaluation! Reach student out on Slack to agree upon evaluation schedule.",
        f":warning: The evaluations was created via {settings.FRONTEND_URL} and is in beta-test. Inform us of any bugs!",
        get_scale_team_info_format_string(scale_team)
    ])
    users = [login for login in scale_team.users]
    users.append(scale_team.evaluator)
    for user in users:
        if scale_team.evaluator == 'moulinette': continue
        try:
            send_messege_to_user(user, message)
        except ValueError as e:
            peersphere_logger.log_error(f"Couldnt send slack message to '{user}'")
            print(f"Failed to get slack id for {user}")

def post_slack_staff_scale_update_success(scale_team:ScaleTeam):
    header = ":scales: Evaluation Successful :scales:"
    message = '\n'.join([
        get_scale_team_info_format_string(scale_team),
        f">*Final Mark:* `{scale_team.final_mark}`"
    ])
    post_message_to_channel(header, message)

def post_slack_staff_scale_update_cancelled(scale_team:ScaleTeam, cancelled_eval_count:int):
    header = ":scales: Evaluation Cancelled :scales:"
    message = '\n'.join([
        get_scale_team_info_format_string(scale_team),
        f">*Cancelled by:* `{scale_team.truant}` (Already cancelled `{cancelled_eval_count}` times.)"
    ])
    post_message_to_channel(header, message)

def post_slack_users_scale_update_cancelled(scale_team:ScaleTeam):
    message = '\n'.join([
        f"The following Evaluation has been cancelled by {scale_team.truant}.",
        get_scale_team_info_format_string(scale_team)
    ])
    users = [login for login in scale_team.users]
    users.append(scale_team.evaluator)
    for user in users:
        if user == 'moulinette': continue
        try:
            send_messege_to_user(user, message)
        except ValueError as e:
            peersphere_logger.log_error(f"Couldnt send slack message to '{user}'")
            print(f"Failed to get slack id for {user}")

def post_slack_location_update(user_login:str, state:bool):
    header = "Location Update"
    if state == True:
        message = f"{user_login} logged in on campus."
    elif state == False:
        message = f"{user_login} logged out from campus."
    post_message_to_channel(header, message)
