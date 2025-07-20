from rest_framework.request import Request
from .enums import WebhookModel, WebhookEvent
from .exceptions import WebhookAuthError, WebhookFormatError
import hmac


class WebhookRequest:
    def __init__(self, _request:Request):
        self.headers:dict = _request.headers
        self.data:dict = _request.data
        self.model:WebhookModel = self._extract_model_from_request()
        self.event:WebhookEvent = self._extract_event_from_request()
        self.secret:str = self._extract_secret_from_request()
        self.request = _request
        pass

    def __str__(self):
        return f"Webhook Request: {self.model}/{self.event} ({self.request.data})"

    def _extract_model_from_request(self) -> WebhookModel:
        model_str:str = self.headers.get('X-Model')
        if model_str == None:
            raise WebhookFormatError("Webhook is missing X-Model header field")
        if model_str in WebhookModel._value2member_map_:
            return WebhookModel(model_str)
        else:
            raise WebhookFormatError(f"Couldn't resolve X-Model type: {model_str}")

    def _extract_event_from_request(self) -> WebhookEvent:
        event_str:str = self.headers.get('X-Event')
        if event_str == None:
            raise WebhookFormatError("Webhook is missing X-Event header field")
        if event_str in WebhookEvent._value2member_map_:
            return WebhookEvent(event_str)
        else:
            raise WebhookFormatError(f"Couldn't resolve X-Event type: {event_str}")

    def _extract_secret_from_request(self) -> str:
        secret = self.headers.get('X-Secret')
        if secret == None:
            raise WebhookFormatError("Webhook is missing X-Secret header field")
        return secret

    def validate(self, _secret:str) -> None:
        if not hmac.compare_digest(_secret, self.secret):
            raise WebhookAuthError()
