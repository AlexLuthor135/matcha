import os, sys
from celery import Celery
from celery.schedules import crontab
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.append(str(BASE_DIR))
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'backend.settings')

app = Celery('backend')
app.config_from_object('django.conf:settings', namespace='CELERY')
app.autodiscover_tasks()

app.conf.beat_schedule = {
    'update-pending-evals-weekly': {
        'task': 'intra_client.tasks.evals.update_pending_evals',
        'schedule': crontab(hour=0, minute=0, day_of_week='mon'),
    },
}
