import os
from django.utils import timezone

class PeersphereLogger():

    LOG_TYPES = [
        "location",
        "quests",
        "evals",
        "usage",
        "error",
        "slack",
        "reset",
        "admins",
        "actions",
        "celery",
    ]

    def __init__(self):
        self._createDirs()

    def _createDirs(self) -> None:
        if not os.path.exists("/backend/logs"):
            os.makedirs("/backend/logs")
        for log_type in self.LOG_TYPES:
            if not os.path.exists(f"/backend/logs/{log_type}"):
                os.makedirs(f"/backend/logs/{log_type}")

    def _get_timestamp(self) -> str:
        timestamp = timezone.now().strftime('%Y-%m-%d %H:%M:%S')
        return timestamp
    
    def _get_date(self) -> str:
        timestamp_day = timezone.now().strftime('%Y-%m-%d')
        return timestamp_day

    def log_location(self, msg:str) -> None:
        with open(f"/backend/logs/location/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")

    def log_eval(self, msg:str) -> None:
        with open(f"/backend/logs/evals/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")

    def log_quest(self, msg:str) -> None:
        with open(f"/backend/logs/quests/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")

    def log_usage(self, msg:str) -> None:
        with open(f"/backend/logs/usage/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")
    
    def log_error(self, msg:str) -> None:
        with open(f"/backend/logs/error/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")

    def log_slack(self, msg:str) -> None:
        with open(f"/backend/logs/slack/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")
    
    def log_reset(self, msg:str) -> None:
        with open(f"/backend/logs/reset/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")
            
    def log_admins(self, msg:str) -> None:
        with open(f"/backend/logs/admins/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")

    def log_action(self, msg:str) -> None:
        with open(f"/backend/logs/actions/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")
            
    def log_celery(self, msg:str) -> None:
        with open(f"/backend/logs/celery/{self._get_date()}", 'a') as log_file:
            log_file.write(f"{self._get_timestamp()} {msg}\n")


peersphere_logger = PeersphereLogger()
