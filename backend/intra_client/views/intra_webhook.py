from rest_framework.views import APIView
from rest_framework.permissions import AllowAny
from rest_framework.response import Response
from rest_framework.request import Request
from rest_framework.exceptions import PermissionDenied
from django.conf import settings
from django.db import transaction
from api.models import User
from .intra_webhooks.exceptions import WebhookAuthError, WebhookFormatError
from .intra_webhooks.enums import WebhookModel
from .intra_webhooks import WebhookRequest, WebhookHandler, WebhookHandlerLocation, WebhookHandlerQuestsUser, WebhookHandlerScaleTeam
import pytz, os
from datetime import datetime

class IntraWebhook(APIView):
    permission_classes = [AllowAny]

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def post(self, request:Request) -> Response:
        try:
            webhook:WebhookRequest = WebhookRequest(request)
            handler:WebhookHandler = self._create_webhook_handler(webhook)
            response:Response = handler.process_and_respond()
            self._log_webhook_interaction(webhook, response)
        except Exception as e:
            response:Response = self._handle_exception_and_respond(e)
            self._log_webhook_interaction(None, response)
        return response
    
    def _create_webhook_handler(self, webhook:WebhookRequest) -> WebhookHandler:
        # instanciate WebkoohHandler based on model
        if webhook.model == WebhookModel.LOCATION:
            return WebhookHandlerLocation(webhook)
        elif webhook.model == WebhookModel.SCALE_TEAM:
            return WebhookHandlerScaleTeam(webhook)
        elif webhook.model == WebhookModel.QUESTS_USER:
            return WebhookHandlerQuestsUser(webhook)
        else:
            return WebhookHandler(webhook)

    def _log_webhook_interaction(self, request:WebhookRequest, response:Response) -> None:
        timestamp = datetime.now(pytz.timezone('Europe/Berlin')).strftime('%Y-%m-%d %H:%M:%S')
        log_str:str = timestamp + " | " + str(request) + f"\n -> Response: {response.status_code} - {response.data}"
        if not os.path.exists("/backend/logs"):
            os.makedirs("/backend/logs")
        with open("/backend/logs/webhook_log", "a") as log_file:
            log_file.write(log_str + "\n")
        pass

    def _handle_exception_and_respond(self, e:Exception) -> Response:
        expected_exceptions = (WebhookAuthError, WebhookFormatError)
        if isinstance(e, expected_exceptions):
            return Response(status=e.status, data={"message": e.message})
        else:
            print(e)
            return Response(status=500, data={"message": "Internal error"})
