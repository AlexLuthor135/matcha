from enum import Enum

class WebhookModel(Enum):
    SCALE_TEAM = 'scale_team'
    LOCATION = 'location'
    QUESTS_USER = 'quests_user'

class WebhookEvent(Enum):
    CREATE = 'create'
    UPDATE = 'update'
    DESTROY = 'destroy'
    CLOSE = 'close'
