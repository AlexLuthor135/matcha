import os
import hmac
import hashlib
import httpx
import json
import logging
from typing import Iterable
from django.conf import settings
from intra_client.services.logging import peersphere_logger

LOG = logging.getLogger(__name__)

def link(login: str):
    """Turns a login into a link for a markdown mesasge on slack"""
    return f'<https://profile.intra.42.fr/users/{login}|{login}>'

def devide_x_by_x(step: int, tab: Iterable):
    login_blocks = []
    array = []
    for name in tab:
        array.append(name)
        if len(array) == step:
            login_blocks.append(array)
            array = []
    
    login_blocks.append(array)
    return login_blocks


def divide_x_by_x_v3(step: int, tab: dict, is_for_first_message: bool):
    login_blocks = []
    array = []
    for login, day in tab.items():
        if is_for_first_message:
            array.append(f'<https://profile.intra.42.fr/users/{login}|{login}> {day}d')
        else:
            array.append(f"<https://profile.intra.42.fr/users/{login}|{login}> {day['old_pace']} -> {day['new_pace']}")
        if len(array) == step:
            login_blocks.append(array)
            array = []
    
    login_blocks.append(array)
    return login_blocks


def validate_slack_signature(signing_secret: str, data: str, timestamp: str, signature: str) -> bool:
    format_req = str.encode(f"v0:{timestamp}:{data}")
    encoded_secret = str.encode(signing_secret)
    request_hash = hmac.new(encoded_secret, format_req,
                            hashlib.sha256).hexdigest()
    calculated_signature = f"v0={request_hash}"

    return hmac.compare_digest(calculated_signature, signature)

def slack_post_message(message: str, channel_id: str, blocks: str = '', slack_bot_token: str = ''):
    headers = {'accept': 'application/json'}
    token_to_use = slack_bot_token if slack_bot_token else settings.SLACK_BOT_TOKEN
    data = {
        'token': token_to_use,
        'channel': channel_id,
        'text': message,
    }

    if blocks != '':
        data['blocks'] = blocks

    res = httpx.post("https://slack.com/api/chat.postMessage",
                     headers=headers, data=data)

    peersphere_logger.log_slack(f"POST: {res.status_code} {res.content}")

    if json.loads(res.content)["ok"] != True:
        print(res.content)

    if res.status_code != 200:
        raise Exception("Error posting message to slack")


def slack_post_ephemeral(message: str, channel_id: str, user_id: str, blocks: str = ''):
    headers = {'accept': 'application/json'}

    data = {
        'token': settings.SLACK_BOT_TOKEN,
        'channel': channel_id,
        'text': message,
        'user': user_id,
        'blocks': blocks
    }

    httpx.post("https://slack.com/api/chat.postEphemeral",
               headers=headers, data=data)


async def slack_get_message(channel: str, ts: str):
    headers = {
        'accept': 'application/json',
    }

    data = {
        "token": settings.SLACK_BOT_TOKEN,
        "channel": channel,
        "oldest": ts,
        "inclusive": 1,
        "limit": 1,
    }

    req = httpx.post("https://slack.com/api/conversations.history",
                     headers=headers, data=data)

    return req


async def slack_delete_message(channel: str, ts: str):
    headers = {
        'accept': 'application/json',
    }

    data = {
        "token": settings.SLACK_BOT_TOKEN,
        "channel": channel,
        "ts": ts,
    }

    httpx.post("https://slack.com/api/chat.delete",
               headers=headers, data=data)


async def slack_update_message(channel: str, ts: str, blocks: str = ''):
    headers = {
        'accept': 'application/json',
    }

    data = {
        "token": settings.SLACK_BOT_TOKEN,
        "channel": channel,
        "ts": ts,
        "blocks": blocks
    }

    httpx.post("https://slack.com/api/chat.update",
               headers=headers, data=data)

def pretty_json(json_data):
    try:
        return json.dumps(json_data, sort_keys=True, indent=4)
    except:
        try:
            data = json.load(json_data)
            return json.dumps(data, sort_keys=True, indent=4)
        except:
            return json_data

def post_message_to_channel(header: str, message: str) -> None:
    slack_post_message("", settings.SLACK_CHANNEL_ID, [[
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": header
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": message
            }
        }
    ]])

def get_slack_user_id_by_email(email: str, slack_bot_token: str):
    with httpx.Client() as client:
        response = client.get(
            'https://slack.com/api/users.lookupByEmail',
            headers={
                'Authorization': f'Bearer {slack_bot_token}',
            },
            params={'email': email}
        )
        result = response.json()
        if result.get('ok'):
            return result['user']['id']
        else:
            LOG.error(f"Failed to retrieve Slack user ID for email {email}: {result.get('error')}")
            return None
