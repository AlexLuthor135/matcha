from intra_client.services.intra_api import ic
from intra_client.views.intra_webhooks import WebhookRequest
from requests import Response

class ScaleTeam:
	def __init__(self, request:WebhookRequest):
		self.id = request.data["id"]
		self._api_obj = self._get_scale_team_from_api(self.id)
		evaluator_dict:dict = self._api_obj["corrector"]
		self.evaluator = evaluator_dict.get("login", "moulinette")
		self.users:list = [user["login"] for user in self._api_obj["correcteds"]]
		self.project_id = self._api_obj["team"]["project_id"]
		self.project_slug = request.data['project']['slug']
		self.final_mark = request.data["final_mark"]
		self.truant = ""
		truant_obj = self._api_obj.get("truant", None)
		if truant_obj != None:
			self.truant = truant_obj.get("login", "")
		for user in self._api_obj["team"]["users"]:
			if user["leader"] == True:
				self.project_user_id = user["projects_user_id"]

	def _get_scale_team_from_api(self, id:str):
		url = f"scale_teams/{id}"
		res:Response = ic.get(url)
		if res.status_code != 200:
			raise Exception()
		return res.json()
	
	def __str__(self):
		return f"{self.id} {self.project_slug} by {self.users} evaluated by {self.evaluator}"
