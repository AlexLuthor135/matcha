from .handler_base import *
from .enums import WebhookEvent
from intra_client.services.logging import peersphere_logger
from django.utils.dateparse import parse_datetime
from django.utils.timezone import is_naive, make_aware
from datetime import datetime, timezone
# from slack.views.views import post_slack_location_update


class WebhookHandlerLocation(WebhookHandler):
    def __init__(self, _request:WebhookRequest):
        super().__init__(_request)
        self._parse_logtimes_from_request()

    def _parse_logtimes_from_request(self):
        self.login_time: str = self.request.data.get("begin_at")
        if self.login_time:
            self.login_time = self.login_time.strip("UTC ")
        self.logout_time:str = self.request.data.get("end_at")
        if self.logout_time:
            self.logout_time = self.logout_time.strip("UTC ")

    def process_and_respond(self) -> Response:
        if self.request.event == WebhookEvent.CREATE:
            return self._handle_create()
        elif self.request.event == WebhookEvent.CLOSE:
            return self._handle_close()
        else:
            return self._get_response_not_implemented()

    def _handle_create(self) -> Response:
        self.request.validate(settings.LOCATION_CREATE)

        user_login = self.request.data["user"]["login"]
        self._update_user_logtimes(user_login)
        peersphere_logger.log_location(f"✅ {user_login} logged in at {self.login_time}")
        return Response({"message": "success"}, status=200)

    def _handle_close(self) -> Response:
        self.request.validate(settings.LOCATION_CLOSE)

        user_login = self.request.data["user"]["login"]
        self._update_user_logtimes(user_login)
        peersphere_logger.log_location(f"🅾️ {user_login} logged out at {self.logout_time}")
        return Response({"status": "success"}, status=200)

    def _update_user_logtimes(self, user_login:str):
        login_time = parse_datetime(self.login_time) if self.login_time else None
        logout_time = parse_datetime(self.logout_time) if self.logout_time else None
        try:
            user = User.objects.get(username=user_login)
            with transaction.atomic():
                if self._date_newer(user.last_campus_login, login_time):
                    user.last_campus_login = login_time
                if self._date_newer(user.last_campus_logout, logout_time):
                    user.last_campus_logout = logout_time
                user.has_location = self._evaluate_login_state(user)
                user.save()
        except User.DoesNotExist:
            print(f"User {user_login} not found in the database.")
        except Exception as e:
            print(f"An error occurred: {e}")

    def _evaluate_login_state(self, user:User) -> bool:
        if not user.last_campus_login:
            return False
        elif not user.last_campus_logout:
            return True
        if is_naive(user.last_campus_login):
            user.last_campus_login = make_aware(user.last_campus_login, timezone.utc)
        if is_naive(user.last_campus_logout):
            user.last_campus_logout = make_aware(user.last_campus_logout, timezone.utc)
        return True if user.last_campus_login > user.last_campus_logout else False

    def _date_newer(self, reference:datetime, date:datetime) -> bool:
        if date is None: return False
        if reference is None: return True
        if is_naive(reference):
            reference = make_aware(reference, timezone.utc)
        if is_naive(date):
            date = make_aware(date, timezone.utc)
        return reference < date
