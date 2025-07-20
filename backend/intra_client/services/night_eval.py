from django.conf import settings
from datetime import datetime
import pytz

def check_night_eval():
    pass
    if not hasattr(settings, 'TIME_CUTOFF_START') or not hasattr(settings, 'TIME_CUTOFF_END'):
        return False
    current_time = datetime.now(pytz.timezone(settings.TIME_ZONE)).time()
    start = settings.TIME_CUTOFF_START
    end = settings.TIME_CUTOFF_END
    if start < end:
        return start <= current_time < end
    else:
        return current_time >= start or current_time < end