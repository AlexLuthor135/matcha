from api.models import Evaluation, User
from celery import shared_task
from intra_client.services.logging import peersphere_logger

@shared_task
def update_pending_evals():
    evals = Evaluation.objects.filter(state=Evaluation.State.PENDING)
    peersphere_logger.log_celery(f"Checking {evals.count()} Evals.")
    for eval in evals:
        if eval.update_from_api():
            peersphere_logger.log_celery(f"Updated {eval.id} from {eval.time_created} -> {eval.state}")

    users = User.objects.filter(has_eval=True)
    peersphere_logger.log_celery(f"Checking {users.count()} users.")
    for user in users:
        if Evaluation.objects.filter(evaluator=user.username, state=Evaluation.State.PENDING).count() == 0:
            user.has_eval = False
            user.save()
            peersphere_logger.log_celery(f"Updated user {user.username}")
