import re
from typing import List, Dict
from slack.services.slack_utils import *

DEFAULT_TEXT_TYPE = 'mrkdwn'

class SlackMessage:
    def __init__(self):
        self.blocks: List[Dict] = []
    
    def add_header(self, content: str, text_type: str = 'plain_text'):
        self.blocks.append({
            "type": "header",
            "text": {
                "type": text_type,
                "text": content,
            }
        })


    def add_text(self, content: str, text_type: str = DEFAULT_TEXT_TYPE,
                    accessory: dict = None):
        if not accessory:
            self.blocks.append({
                    "type": "section",
                    "text": {
                        "type": text_type,
                        "text": content,
                    }
                })
        else:
            self.blocks.append({
                    "type": "section",
                    "text": {
                        "type": text_type,
                        "text": content,
                    },
                    "accessory": accessory
                })


    def add_login_list(self, logins: List[str]):
        list_of_logins = devide_x_by_x(42, logins)

        for li in list_of_logins:
            self.add_text(f'`{", ".join(map(lambda x: link(x), li))}`', text_type='mrkdwn')


    def add_login_list_v3(self, logins: dict, is_for_first_message: bool):
        list_of_logins = divide_x_by_x_v3(42, logins, is_for_first_message)

        for li in list_of_logins:
            self.add_text(f'`{", ".join(li)}`', text_type='mrkdwn')


    def add_divider(self):
        self.blocks.append({
            'type': 'divider'
        })


    def add_context(self, content: str, text_type: str = DEFAULT_TEXT_TYPE):
        self.blocks.append({
            'type': 'context',
            'elements': [
                {
                    'type': text_type,
                    'text': content,
                }
            ]
        })


    def add_block(self, block):
        self.blocks.append(block)

    def send(self, send_to: str, ephemeral_user_id: str = None, display: bool = False):
        blocks_to_send = []
        current_block = []
        match_login_list = r'`(<https:\/\/profile\.intra\.42\.fr\/users\/[^|]*\|[^>]*>,\s)*<https:\/\/profile\.intra\.42\.fr\/users\/[^|]*\|[^>]*>`'

        def is_more_than_threethousand_char(current_block: list, block_to_add: dict):
            count = 0
            
            for b in current_block:
                has_a_text = 'text' in b and 'text' in b['text']
                if has_a_text:
                    count += len(b['text']['text'])
            
            if 'text' in block_to_add and 'text' in block_to_add['text'] \
                and count + len(block_to_add['text']['text']) >= 3000:
                return True
            return False

        for b in self.blocks:
            has_a_text = 'text' in b and 'text' in b['text']
            test_is_login_lists = False if not has_a_text else re.match(match_login_list, b["text"]["text"])

            if is_more_than_threethousand_char(current_block, b):
                blocks_to_send.append(current_block)
                blocks_to_send.append([b])
                current_block = []
            else:
                current_block.append(b)
        
        blocks_to_send.append(current_block)

        if ephemeral_user_id:
            for b in blocks_to_send:
                print(b) if display else None
                slack_post_ephemeral('', send_to,
                    ephemeral_user_id, blocks=str(b))
        else:
            for b in blocks_to_send:
                if len(b):
                    print(pretty_json(b)) if display else None
                    slack_post_message('', send_to, str(b))