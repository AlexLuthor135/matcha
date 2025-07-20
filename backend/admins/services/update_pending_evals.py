from django.core.management import call_command
from rest_framework.response import Response

def update_pending_evals():
	call_command('update_pending_evals')
	return Response({"message": "Evals have been updated."})
