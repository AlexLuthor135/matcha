from intra_client.services.intra_api import ic
import requests

def get_corrector_score(student_login, as_who):
    """Get the corrector score for a student during Common Core.
    'as_who' can be 'as_corrector' or 'as_corrected'."""
    assert as_who in ('as_corrector', 'as_corrected'), "as_who must be 'as_corrector' or 'as_corrected'"
    piscine_project_id = [1255, 1256, 1305, 1257, 1258, 1259, 1260, 1261, 1262, 1263, 1270, 1264, 1265, 1266, 1267, 1268, 1271, 2315, 1308, 1310, 1309,]
    score = 0
    try:
        scale_teams = ic.pages_threaded(f"users/{student_login}/scale_teams/{as_who}")
        for scale_team in scale_teams:
            if scale_team.get('truant') or scale_team.get("team", {}).get("project_id") in piscine_project_id:
                continue
            score += 1
    except requests.exceptions.HTTPError as e:
        if e.response.status_code == 404:
            print(f"User {student_login} not found.")
        else:
            print(f"Error: {e}")
    except Exception as e:
        print(f"Error: {e}")
    return int(score)