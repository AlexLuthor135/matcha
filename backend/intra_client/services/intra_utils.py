""" These are generic functions that can be used to get data from the API.
    The aim is to avoid repeating code and add here generic functions that will
    return data in a format that is easy to use.
"""
import json
import string

from datetime import datetime, timedelta
from intra_client.services.intra_api import IntraAPIClient
from django.conf import settings


def get_current_students(ic: IntraAPIClient, campus: int):
    """ Returns a list of all active students in a given campus.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            campus (int): id of the campus, e.g. 51 for Berlin, 44 for Wolfsburg

        Returns:
            users (list): list of users
    """
    endpoint = f"/cursus/21/cursus_users"

    payload = {
        "filter[campus_id]": str(campus),
        # "range[begin_at]": "2025-01-01T00:00:00.000Z,2025-06-01T00:00:00.000Z"
    }

    users = ic.pages_threaded(url=endpoint, params=payload)
    filtered_users = []
    for user in users:
        if user.get("end_at") is None and user.get("staff?") is not True:
            filtered_users.append(user)
        else:
            print(f'User {user.get("login")} was removed')
    # for user in users:
    #     if user['pool_year'] == '2019':
    #         users.remove(user)

    return filtered_users


def get_cursus_users(ic: IntraAPIClient, campus: int, cursus: int):
    """ Returns a list of all students in a given campus and cursus.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            campus (int): id of the campus, e.g. 51 for Berlin, 44 for Wolfsburg
            cursus (int): id of the campus, e.g. 21 for 42cursus

        Returns:
            users (list): list of users
    """
    endpoint = f"/cursus/{str(cursus)}/cursus_users"

    payload = {
        "filter[primary_campus_id]": str(campus),
    }

    users = ic.pages_threaded(url=endpoint, params=payload)

    return users


# def get_unique_logins_day(ic: IntraAPIClient, campus: str, day_offset: int):
#     day = (datetime.now() - timedelta(days=day_offset)).strftime('%Y-%m-%d')

#     endpoint = f"/campus/{campus}/locations"

#     payload = {
#         "range[begin_at]": f"{day}T00:00:00.000Z,{day}T23:59:59.000Z"
#     }

#     locations = ic.pages_threaded(url=endpoint, params=payload)

#     unique_users = set()
#     for location in locations:
#         unique_users.add(location['user']['login'])

#     return len(unique_users)

def get_unique_logins_day(ic: IntraAPIClient, campus: str, day_offset: int):
    day = (datetime.now() - timedelta(days=day_offset)).strftime('%Y-%m-%d')

    endpoint = f"/campus/{campus}/locations"

    payload = {
        "range[begin_at]": f"{day}T00:00:00.000Z,{day}T23:59:59.000Z"
    }

    locations = ic.pages_threaded(url=endpoint, params=payload)

    try:
        students = set()
        pisciners = set()
        for location in locations:
            if location['user']['login'] in students or \
                    location['user']['login'] in pisciners:
                continue

            user_info = ic.get(f"users/{location['user']['id']}").json()
            if len(user_info['cursus_users']) == 1:
                pisciners.add(location['user']['login'])
                # print('pisciner: ', location['user']['login'])
            else:
                students.add(location['user']['login'])
                # print('student: ', location['user']['login'])

    except Exception as e:
        print(e)

    return len(students), len(pisciners)


def get_concurrent_logins(ic: IntraAPIClient, campus: str, day_offset: int, step: int = 1):
    return 0
    day = (datetime.now() - timedelta(days=day_offset)).strftime('%Y-%m-%d')

    endpoint = f"/campus/{campus}/locations"

    concurrent_logins = {
        'value': 0,
        'time_range': '',
    }

    for hour in range(0, 24, step):
        payload = {
            "range[begin_at]": f"{day}T{str(hour).zfill(2)}:00:00.000Z,{day}T{str(hour).zfill(2)}:59:59.000Z"
        }

        locations = ic.pages_threaded(url=endpoint, params=payload)

        unique_users = set()
        for location in locations:
            unique_users.add(location['user']['login'])

        if len(unique_users) > concurrent_logins['value']:
            concurrent_logins['value'] = len(unique_users)
            concurrent_logins['time_range'] = f"{str(hour).zfill(2)}:00 - {str(hour + step).zfill(2)}:00"

    return concurrent_logins


def get_pool_balance(ic: IntraAPIClient, pool_id: int):
    """ Returns the current balance of a given pool.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            pool_id (int): id of the pool

        Returns:
            balance (obj): current and max points of the pool
    """
    endpoint = f"/pools/{str(pool_id)}"

    pool = ic.get(url=endpoint)

    return pool


def get_future_blackholes(ic: IntraAPIClient, days_from: int, days_to: int):
    """ Returns a list of all students that will be blackholed in a time range
        between days_from and days_to.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            days_from (int): start date in days from now
            days_to (int): end date in days from now

        Returns:
            users (list): list of users
    """
    endpoint = f"/cursus/21/cursus_users"

    start_date = datetime.now() + timedelta(days=days_from)
    end_date = datetime.now() + timedelta(days=days_to)

    start_date = datetime(
        start_date.year, start_date.month, start_date.day, 0, 0, 0)
    end_date = datetime(end_date.year, end_date.month,
                        end_date.day, 23, 59, 59)

    payload = {
        "filter[campus_id]": f"{settings.CAMPUS_ID}",
        "range[blackholed_at]": f"{start_date},{end_date}",
    }

    users = ic.pages_threaded(url=endpoint, params=payload)

    for user in users:
        if user["user"]["staff?"] is True or \
           user["user"]["pool_year"] == "2019" or \
           user["user"]["kind"] != "student":
            users.remove(user)

    users.sort(key=lambda x: x["user"]["login"])

    return users


def get_users_from_piscine(ic: IntraAPIClient, campus: string, month: string, year: string, created_at: str = None):
    """ Returns a list of students doing a piscine in a given campus, month and
        year.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            campus (string): name/id of the campus, e.g. berlin/51, wolfsburg/44
            month (string): name of the month, e.g. july
            year (string): year of the piscine, e.g. 2022

        Returns:
            users (list): list of users
    """
    endpoint = f"/campus/{campus}/users"

    payload = {
        "filter[pool_year]": year,
        "filter[pool_month]": month,
        "filter[kind]": "student",
    }

    if created_at:
        payload["range[created_at]"] = created_at

    users = ic.pages_threaded(url=endpoint, params=payload)

    return users


def get_users_with_active_location(ic: IntraAPIClient, campus: string):
    """ Returns a list of students with an active location in a given campus.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            campus (string): name/id of the campus, e.g. berlin/51, wolfsburg/44

        Returns:
            users (list): list of users
    """
    endpoint = f"/campus/{campus}/locations"

    payload = {
        "filter[active]": "true"
    }

    locations = ic.pages_threaded(url=endpoint, params=payload)

    return locations


def delete_location(ic: IntraAPIClient, location_id: string) -> None:
    """Delete a location with a given id.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            location_id (string): id of the location
    """
    endpoint = f"/locations/{location_id}"
    ic.delete(endpoint)


def ask_piscine_date():
    """Asks the user for the piscine date and returns the month, month number
        and year as a tuple.
    """
    valid_months = ['january', 'february', 'march', 'april', 'may', 'june',
                    'july', 'august', 'september', 'october', 'november', 'december']

    month = input('🗳 Enter month of target piscine (e.g. july): ')
    while (month not in valid_months):
        month = input('🗳 Enter month of target piscine (e.g. july): ')
    month_number = valid_months.index(month) + 1

    year = input('🗳 Enter year of target piscine (e.g. 2022): ')
    while (int(year) < 2000 or int(year) > 2100):
        year = input('🗳 Enter year  of target piscine (e.g. 2022): ')

    return month, month_number, year


def ask_rush_project_id() -> str:
    """Asks the user for a rush project id and returns its ID.
    """
    rushes_ids = ["1308", "1310", "1309"]

    print("Select a Rush:")
    print("0) Rush 00")
    print("1) Rush 01")
    print("2) Rush 02")

    idx = int(input('> '))
    if (idx < 0 or idx > 2):
        print("Invalid project")
        exit(1)

    return rushes_ids[idx]


def get_project_teams(ic: IntraAPIClient, project_id: string):
    """Get all teams for a project given its id.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            project_id (string): id of the project

        Returns:
            teams (list): list of teams
    """
    endpoint = f"projects/{project_id}/teams"

    payload = {
        "filter[campus]": settings.CAMPUS_ID,
        "filter[with_mark]": "false"
    }

    teams = ic.pages_threaded(url=endpoint, params=payload)

    return teams


def get_users_doing_project(ic: IntraAPIClient, project_id: string, range_start: string = '', range_end: string = ''):
    """ Get all users doing a project given its id in an optional start date range.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            project_id (string): id of the project
            range_start (string): start date of the range in format YYYY-MM-DD
            range_end (string): end date of the range in format YYYY-MM-DD

        Returns:
            users (list): list of users
    """
    endpoint = f"/projects/{project_id}/projects_users"

    payload = {
        "filter[campus]": settings.CAMPUS_ID,
        "filter[marked]": "false"
    }

    if (range_start != '' and range_end != ''):
        payload['range[created_at]'] = f"{range_start}T00:00:00.000Z,{range_end}T00:00:00.000Z"

    users = ic.pages_threaded(url=endpoint, params=payload)
    return users


def get_user_info(ic: IntraAPIClient, user_id: string, get_candidature: bool = False):
    """ Get user info given its id and optionally include candidature info.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            user_id (string): id of the user
            get_candidature (bool): whether to include candidature info

        Returns:
            user (dict): user info
    """
    endpoint = f"users/{user_id}"

    user = ic.get(url=endpoint).json()

    if get_candidature:
        candidature_info = get_candidature_info(ic, user_id)
        for key, value in candidature_info.items():
            user[key] = value

    return user


def get_candidature_info(ic: IntraAPIClient, user_id: string):
    """ Get candidature info given a user id.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            user_id (string): id of the user

        Returns:
            user (dict): candidature info
    """
    endpoint = f"users/{user_id}/user_candidature"

    user = ic.get(url=endpoint).json()

    return user


def delete_team(ic: IntraAPIClient, team_id: string):
    """ Delete a team given its team id.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            team_id (string): id of the team

        Returns:
            bool: whether the team was deleted
    """
    endpoint = f"/teams/{team_id}"

    delete = ic.delete(url=endpoint)

    if delete.status_code == 204:
        return True
    return False


def patch_location(ic: IntraAPIClient, location_id: string, end_at: string):
    """ Patch a location given its id and a payload.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            location_id (string): id of the location
            end_at (string): end date of the location in format YYYY-MM-DD

        Returns:
            bool: whether the location was patched
    """
    endpoint = f"/locations/{location_id}"

    payload = {
        "location[end_at]": f"{end_at}"
    }

    req = ic.patch(url=endpoint, data=payload)

    if req.status_code == 204:
        return True
    return False


def post_tig(ic: IntraAPIClient, user_id: string, tiger: string, reason: string, duration: string):
    headers = {
        'Content-Type': 'application/json',
    }

    payload = json.dumps({
        "close": {
            "closer_id": f"{tiger}",
            "kind": "other",
            "reason": f"{reason}",
            "community_services_attributes": [
                {
                    "duration": f"{duration}",
                    "occupation": "Ask at front desk"
                }
            ]
        }
    })

    ic.post(f'users/{user_id}/closes', data=payload, headers=headers)


def create_evaluation(ic: IntraAPIClient, evaluator_id: string, team_id: string, begin_at: string):
    headers = {
        'Content-Type': 'application/json',
    }

    # "begin_at": "2022-07-11T10:17:40,090",
    payload = json.dumps({
        "scale_teams": [{
            "team_id": int(team_id),
            "begin_at": begin_at,
            "user_id": int(evaluator_id)
        }]
    })

    req = ic.post('scale_teams/multiple_create', data=payload, headers=headers)
    return req.status_code == 201


def get_events_feedback(ic: IntraAPIClient, campus_id: int, cursus_id: int, range_start: string = '', range_end: string = ''):
    endpoint = f"campus/{campus_id}/cursus/{cursus_id}/events"

    payload = {
        "filter[future]": "false",
    }

    if (range_start != '' and range_end != ''):
        payload['range[end_at]'] = f"{range_start}T00:00:00.000Z,{range_end}T23:59:59.999Z"

    events = ic.pages_threaded(url=endpoint, params=payload)

    for event in events:
        if len(event['cursus_ids']) > 1 and cursus_id == 9:
            events.remove(event)

    payload = {
        "filter[campus_id]": campus_id,
    }

    feedbacks = []
    for event in events:
        endpoint = f"events/{event['id']}/feedbacks"
        req = ic.pages_threaded(url=endpoint, params=payload)

        feedbacks.append({
            "name": event['name'],
            "begin_at": event['begin_at'],
            "id": event['id'],
            "feedbacks": req
        })

    feedbacks.sort(key=lambda x: x['begin_at'])

    return feedbacks

def get_quest_users(ic: IntraAPIClient, campus: string, quest_id: int, created_at: str = None, validated_at: bool = False):
    """ Returns a list of students in specified rank.

        Parameters:
            ic (IntraAPIClient): instance of the IntraAPIClient
            campus (string): name/id of the campus, e.g. berlin/51, wolfsburg/44
            quest (int): id of the quest rank
                Common Core Rank 00 - ID: 44
                Common Core Rank 01 - ID: 45
                Common Core Rank 02 - ID: 46
                Common Core Rank 03 - ID: 47
                Common Core Rank 04 - ID: 48
                Common Core Rank 05 - ID: 49
                Common Core Rank 06 - ID: 59
                Common Core - ID: 37

        Returns:
            users (list): list of users
    """
    endpoint = f"/quests/{quest_id}/quests_users"

    payload = {
        "filter[campus_id]": campus,
    }

    if created_at:
        payload["range[created_at]"] = created_at
    

    quest_users = ic.pages_threaded(url=endpoint, params=payload)
    users = []

    if quest_id == 37:
        if validated_at:
            validated_users = []
            for user in quest_users:
                if user.get('validated_at'):
                    validated_users.append(user)
            quest_users = validated_users
        else:
            not_validated_users = []
            for user in quest_users:
                if not user.get('validated_at'):
                    not_validated_users.append(user)
            quest_users = not_validated_users

    for user in quest_users:
        users.append(user["user"])

    return users

def get_users_quest_rank(ic: IntraAPIClient, user_id: int):
    """Returns the rank  of the user depending on the last quest_user
                Common Core Rank 00 - ID: 44
                Common Core Rank 01 - ID: 45
                Common Core Rank 02 - ID: 46
                Common Core Rank 03 - ID: 47
                Common Core Rank 04 - ID: 48
                Common Core Rank 05 - ID: 49
                Common Core Rank 06 - ID: 59
                Common Core - ID: 37
    """
    quest_users = ic.pages_threaded(f"users/{user_id}/quests_users")
    last_quest_id = 0

    for quest_user in quest_users:
        if quest_user.get("quest_id") == 37:
            if quest_user.get("validated_at"):
                return 7
        elif quest_user.get("quest_id") > last_quest_id:
            last_quest_id = quest_user.get("quest_id")
    rank_mapping = {
        0: 0,
        44: 1,
        45: 2,
        46: 3,
        47: 4,
        48: 5,
        49: 6,
        59: 6,
    }
    return rank_mapping.get(last_quest_id, last_quest_id)
