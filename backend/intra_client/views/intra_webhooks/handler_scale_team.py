from django.utils import timezone
from .handler_base import *
from .enums import WebhookEvent
from intra_client.services.scale_team_class import ScaleTeam
from intra_client.services.logging import peersphere_logger
from slack.views.views import post_slack_scale_create, post_slack_scale_update_cancelled, post_slack_staff_scale_update_success

INTERNSHIP_IDS = [1665, 1666, 1641, 1642, 1647, 1648, 1654, 1653, 1659, 1660,]

class WebhookHandlerScaleTeam(WebhookHandler):
    def __init__(self, _request:WebhookRequest):
        super().__init__(_request)
        if self.request.event == WebhookEvent.DESTROY:
            self.scale_team_obj:ScaleTeam = None
            return
        try:
            self.scale_team_obj:ScaleTeam = ScaleTeam(self.request)
        except Exception as e:
            print(e)
            self.scale_team_obj:ScaleTeam = None
    
    def process_and_respond(self) -> Response:
        if self.request.event == WebhookEvent.DESTROY:
            self.request.validate(settings.SCALE_TEAM_DESTROY)
            result = self._scale_team_destroy()
            return Response(result, status=200)
        if self.scale_team_obj == None:
            return Response({"error":"Could not fetch ScaleTeam from api."}, status=500)
        if self.request.event == WebhookEvent.CREATE:
            self.request.validate(settings.SCALE_TEAM_CREATE)
            print(f'I created a new team: {self.request.data["team"]["name"]}')
            post_slack_scale_create(self.scale_team_obj)
            result = self._scale_team_create()
            return Response(result, status=200)
        elif self.request.event == WebhookEvent.UPDATE:
            self.request.validate(settings.SCALE_TEAM_UPDATE)
            print(f'I updated the team: {self.request.data["team"]["name"]}')
            result = self._scale_team_update()
            return Response(result, status=200)
        else:
            return self._get_response_not_implemented()

    def _scale_team_create(self) -> dict:
        """
        Makes the user in a database as having an evaluation,
        and not being able to evaluate other students
        """
        evaluator_login = self.scale_team_obj.evaluator
        project_id = self.request.data["team"]["project_id"]

        #Project IDs for internship projects. They are always pending, therefore should be ignored.
        if project_id in INTERNSHIP_IDS:
            return {"message": "Internship project detected. Ignoring."}
        try:
            user = User.objects.get(username=evaluator_login)
            with transaction.atomic():
                user.has_eval = True
                user.save()
                print(f"User {evaluator_login} has been marked as having an evaluation.")
                self._create_evaluation_in_db()
                self._reduce_score_for_team(self.scale_team_obj.users)
                return {"message": f"User {evaluator_login} has been marked as having an evaluation."}
        except User.DoesNotExist:
            print(f"User {evaluator_login} not found in the database.")
            return {"error": f"User {evaluator_login} not found in the database."}
        except Exception as e:
            print(f"An error occurred: {e}")
            return {"error": f"An error occurred: {str(e)}"}
    
    def _reduce_score_for_team(self, team:list) -> None:
        for login in team:
            try:
                user = User.objects.get(username=login)
                with transaction.atomic():
                    user.badpeer_score -= 1
                    user.save()
                print(f"Deducted peer score for {login}")
            except User.DoesNotExist:
                print(f"User {login} not found in the database.")
            except Exception as e:
                print(f"An error occurred: {e}")

    def _scale_team_update(self) -> dict:
        """
        Checks whether the evaluation was successful or cancelled.
        """
        #Project IDs for internship projects. They are always pending, therefore should be ignored.
        if self.request.data["team"]["project_id"] in INTERNSHIP_IDS:
            return {"message": "Internship project detected. Ignoring."}
        #user_login = hard to get, maybe important
        if self.scale_team_obj.truant == "":
            return self._scale_team_update_success()
        else:
            return self._scale_team_update_cancelled()

    def _scale_team_update_success(self) -> dict:
        try:
            evaluator = User.objects.get(username=self.scale_team_obj.evaluator)
            with transaction.atomic():
                evaluator.has_eval = False
                evaluator.happeer_score += 1
                evaluator.badpeer_score += 1
                evaluator.save()
                print(f"User {evaluator} has been marked as not having an evaluation.")
            self._update_friend_lists()
            self._update_evaluation_in_db(Evaluation.State.SUCCESSFUL)
            post_slack_staff_scale_update_success(self.scale_team_obj)
            peersphere_logger.log_eval(f"✅ {self.scale_team_obj.id} Team: {self.scale_team_obj.users}({self.scale_team_obj.project_slug}) Evaluator:{self.scale_team_obj.evaluator}")
            return {"message": f"User {self.scale_team_obj.evaluator} has been marked as not having an evaluation."}
        except User.DoesNotExist:
            print(f"User {self.scale_team_obj.evaluator} not found in the database.")
            return {"error": f"User {self.scale_team_obj.evaluator} not found in the database."}
        except Exception as e:
            print(f"An error occurred: {e}")
            return {"error": f"An error occurred: {str(e)}"}
        
    def _update_friend_lists(self) -> None:
        for student in self.scale_team_obj.users:
            try:
                user = User.objects.get(username=student)
                with transaction.atomic():
                    if self.scale_team_obj.evaluator not in user.friends:
                        user.friends.append(self.scale_team_obj.evaluator)
                    if len(user.friends) > int(settings.FRIEND_LIST_SIZE):
                        user.friends.pop(0)
                    user.save()
                print(f"Friend added to {student}: {self.scale_team_obj.evaluator}")
            except User.DoesNotExist:
                print(f"User {student} not found in the database.")
            except Exception as e:
                print(f"An error occurred: {e}")

    def _scale_team_update_cancelled(self) -> dict:
        users:list = self.scale_team_obj.users
        evaluator_login = self.scale_team_obj.evaluator
        truant = self.scale_team_obj.truant
        cancelled_eval_count = 0
        try:
            evaluator = User.objects.get(username=evaluator_login)
            with transaction.atomic():
                evaluator.has_eval = False
                evaluator.save()
            user = User.objects.get(username=truant)
            with transaction.atomic():
                user.cancelled_evals += 1
                cancelled_eval_count = user.cancelled_evals
                user.save()
                print(f"User {truant} has cancelled an evaluation.")
            if truant == evaluator_login:
                self._refund_score_for_team(self.scale_team_obj.users)
            self._update_evaluation_in_db(Evaluation.State.CANCELLED)
            peersphere_logger.log_eval(f"❌ {self.scale_team_obj.id} Cacelled by:{truant} Team: {self.scale_team_obj.users}({self.scale_team_obj.project_slug}) Evaluator:{self.scale_team_obj.evaluator}")
            try:
                post_slack_scale_update_cancelled(self.scale_team_obj, cancelled_eval_count)
            except:
                peersphere_logger.log_error(f"Couldn't send slack messages for cancellation")
            return {"message": f"User {truant} has been cancelling an evaluation."}
        except User.DoesNotExist:
            print(f"User {truant} not found in the database.")
            return {"error": f"User {truant} not found in the database."}
        except Exception as e:
            print(f"An error occurred: {e}")
            return {"error": f"An error occurred: {str(e)}"}
        return {"message": "evaluation got cancelled"}

    def _refund_score_for_team(self, team:list) -> None:
        for login in team:
            try:
                user = User.objects.get(username=login)
                with transaction.atomic():
                    user.badpeer_score += 1
                    user.save()
                print(f"Refunded peer score for {login}")
            except User.DoesNotExist:
                print(f"User {login} not found in the database.")
            except Exception as e:
                print(f"An error occurred: {e}")
    
    def _scale_team_destroy(self) -> dict:
        try:
            evaluation = Evaluation.objects.get(id=self.request.data['id'])
            evaluation.state = Evaluation.State.DESTROYED
            evaluation.time_finished = timezone.now()
            evaluation.save()
            evaluator = User.objects.get(username=evaluation.evaluator)
            evaluator.has_eval = False
            evaluator.save()
            peersphere_logger.log_eval(f"🧨 {evaluation}")
        except Evaluation.DoesNotExist:
            peersphere_logger.log_error(f"From SCALE_TEAM/DESTROY: Coudn't find eval in DB")
            print(f"Error: Doesn't exist in DB: {self.scale_team_obj}")
        except User.DoesNotExist:
            peersphere_logger.log_error(f"From SCALE_TEAM/DESTROY: Coudn't find user '{evaluation.evaluator}' in DB")
            print(f"Error: Doesn't exist in DB: {evaluation.evaluator}")
        except Exception as e:
            peersphere_logger.log_error(f"From SCALE_TEAM/DESTROY: {e} - {str(e)}")
            print(f"An error occurred: {e}")
        return {"message": "evaluation got destroyed"}

    def _create_evaluation_in_db(self) -> None:
        if Evaluation.objects.filter(id=self.scale_team_obj.id).first():
            print(f"{self.scale_team_obj} already in database.")
            return
        Evaluation.objects.create(
            id = self.scale_team_obj.id,
            team = self.scale_team_obj.users,
            team_str = ','.join(self.scale_team_obj.users),
            evaluator = self.scale_team_obj.evaluator,
            project = self.scale_team_obj.project_slug
        )
        print(f"Database created: {self.scale_team_obj}")
        peersphere_logger.log_eval(f"🚀 {self.scale_team_obj.id} Team: {self.scale_team_obj.users}({self.scale_team_obj.project_slug}) Evaluator:{self.scale_team_obj.evaluator}")

    def _update_evaluation_in_db(self, state:Evaluation.State) -> None:
        try:
            evaluation = Evaluation.objects.get(id=self.scale_team_obj.id)
            evaluation.state = state
            if state == Evaluation.State.SUCCESSFUL:
                evaluation.final_mark = self.scale_team_obj.final_mark
            evaluation.cancelled_by = self.scale_team_obj.truant
            evaluation.time_finished = timezone.now()
            evaluation.save()
        except Evaluation.DoesNotExist:
            print(f"Error: Doesn't exist in DB: {self.scale_team_obj}")
        except Exception as e:
            print(f"An error occurred: {e}")

