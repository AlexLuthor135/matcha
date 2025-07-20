from .handler_base import *
from .enums import WebhookEvent
from intra_client.services.logging import peersphere_logger

class WebhookHandlerQuestsUser(WebhookHandler):
    def __init__(self, _request:WebhookRequest):
        super().__init__(_request)
    
    def process_and_respond(self) -> Response:
        if self.request.event == WebhookEvent.CREATE:
            self.request.validate(settings.QUESTS_USER_CREATE)
            result = self._quests_user_create()
            return Response(result, status=200)
        elif self.request.event == WebhookEvent.UPDATE:
            self.request.validate(settings.QUESTS_USER_UPDATE)
            result = self._quests_user_update()
            return Response(result, status=200)
        else:
            return self._get_response_not_implemented()


    def _quests_user_create(self) -> dict:
        """
        Creates the user's rank based on the quest_id.

        Returns:
            dict: Status message indicating the result of the operation.
        """
        user_login = self.request.data["user"]["login"]
        quest_id = self.request.data["quest_id"]
        quest_to_rank = {
            37: 0,
            44: 1,
            45: 2,
            46: 3,
            47: 4,
            48: 5,
            49: 6,
        }
        rank_number = quest_to_rank.get(quest_id)
        if rank_number is None:
            return {"error": f"Quest ID {quest_id} not mapped to any rank"}
        try:
            user = User.objects.get(username=user_login)
            with transaction.atomic():
                if user.quest_rank == rank_number:
                    print(f"User {user_login} is already in Rank {rank_number}.")
                    return {"message": f"User {user_login} is already in Rank {rank_number}."}
                user.quest_rank = rank_number
                user.save()
                print(f"User {user_login} moved to Rank {rank_number}.")
                peersphere_logger.log_quest(f"User {user_login} moved to Rank {rank_number}.")
                return {"message": f"User {user_login} moved to Rank {rank_number}."}
        except User.DoesNotExist:
            print(f"User {user_login} not found in the database.")
            return {"error": f"User {user_login} not found in the database."}
        except Exception as e:
            print(f"An error occurred: {e}")
            return {"error": f"An error occurred: {str(e)}"}
        
    def _quests_user_update(self) -> dict:
        """
        Updates the user's rank based on the quest_id.

        Args:
            user_login (str): The login of the user to update.
            quest_id (int): The ID of the quest to determine the new rank.

        Returns:
            dict: Status message indicating the result of the operation.
        """
        user_login = self.request.data["user"]["login"]
        quest_id = self.request.data["quest_id"]
        quest_to_rank = {37: 7}
        rank_number = quest_to_rank.get(quest_id)
        if rank_number is None:
            return {"error": f"Quest ID {quest_id} not mapped to any rank"}
        try:
            user = User.objects.get(username=user_login)
            with transaction.atomic():
                if user.quest_rank == rank_number:
                    print(f"User {user_login} is already in Rank {rank_number}.")
                    return {"message": f"User {user_login} is already in Rank {rank_number}."}
                user.quest_rank = rank_number
                user.save()
                print(f"User {user_login} moved to Rank {rank_number}.")
                peersphere_logger.log_quest(f"User {user_login} moved to Rank {rank_number}.")
                return {"message": f"User {user_login} moved to Rank {rank_number}."}
        except User.DoesNotExist:
            print(f"User {user_login} not found in the database.")
            return {"error": f"User {user_login} not found in the database."}
        except Exception as e:
            print(f"An error occurred: {e}")
            return {"error": f"An error occurred: {str(e)}"}
