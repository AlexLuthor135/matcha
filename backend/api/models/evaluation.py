from django.db import models
from intra_client.services.intra_api import ic
from requests import Response
from intra_client.services.logging import peersphere_logger
from django.utils.dateparse import parse_datetime

class Evaluation(models.Model):
    class State(models.TextChoices):
        PENDING = 'pending'
        SUCCESSFUL = 'successful'
        CANCELLED = 'cancelled'
        DESTROYED = 'destroyed'

    id = models.BigIntegerField(primary_key=True)
    time_created = models.DateTimeField(auto_now_add=True)
    team = models.JSONField(default=list, blank=True)  # Stores a list of team member names
    team_str = models.CharField(max_length=100, null=True)
    evaluator = models.CharField(max_length=10)
    project = models.CharField(max_length=200)
    state = models.CharField(max_length=20, choices=State.choices, default=State.PENDING)
    cancelled_by = models.CharField(max_length=10, blank=True, default='')
    time_finished = models.DateTimeField(blank=True, null=True)
    final_mark = models.IntegerField(null=True, blank=True, default=None)

    def __str__(self):
        return f"Evaluation of {self.project} by {self.team} evaluated by {self.evaluator} ({self.state})"

    def update_from_api(self) -> bool:
        try:
            res: Response = ic.get(f'/scale_teams/{self.id}')
        except ValueError as e:
            if '404' in e.args[0]:
                self.state = self.State.DESTROYED
                self.time_finished = self.time_created
                self.save()
                return True
        res_json = res.json()
        final_mark = res_json.get("final_mark")
        truant = ""
        truant_obj = res_json.get("truant", None)
        if truant_obj != None:
            truant = truant_obj.get("login", "")
        if truant != "":
            self.cancelled_by = truant
            self.state = self.State.CANCELLED
            self.time_finished = parse_datetime(res_json.get("updated_at"))
            self.save()
            return True
        elif final_mark != None:
            self.final_mark = final_mark
            self.state = self.State.SUCCESSFUL
            self.time_finished = parse_datetime(res_json.get("filled_at"))
            self.save()
            return True
        return False

