class RankNode:
    def __init__(self, rank, users):
        self.rank = rank
        self.users = users
        self.prev = None
        self.next = None