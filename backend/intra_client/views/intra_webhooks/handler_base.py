from rest_framework.response import Response
from django.conf import settings
from api.models import User, Evaluation
from django.db import transaction
from .request import WebhookRequest

class WebhookHandler:
    def __init__(self, request:WebhookRequest):
        self.request:WebhookRequest = request

    def process_and_respond(self) -> Response:
        return Response(data={"message":f"No WebhookHandler implemented for the request. ({str(self.request)})"}, status=404)

    def _get_response_not_implemented(self) -> Response:
        response:Response = Response(status=404)
        response.data = {"message": "Webhook not implemented"}
        return response
