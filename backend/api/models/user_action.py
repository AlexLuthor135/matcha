from django.db import models

class UserAction(models.Model):
    class Action(models.TextChoices):
        BLOCK = 'block'
        UNBLOCK = 'unblock'

    id = models.BigAutoField(primary_key=True)
    user = models.CharField(max_length=10)
    action = models.CharField(max_length=20, choices=Action.choices, null=False)
    time = models.DateTimeField(auto_now_add=True)

    def __str__(self):
        return f"{self.action} by {self.user} at {self.time}"
