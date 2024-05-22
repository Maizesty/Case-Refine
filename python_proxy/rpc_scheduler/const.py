from enum import Enum, unique

@unique
class Scheduler(Enum):
    R2_SCHEDULER = "round-robin"
    COMMUNITY_AWARE_SCHEDULER = "commiunity-aware"
    COMMUNITY_AWARE_NO_LB_SCHEDULER = "commiunity-aware-no-lb"
    NAIVE_CACHE_SCHEDULER = "NaiveCache"
Modes = (
    "round-robin",
    "commiunity-aware",
    "commiunity-aware-no-lb",
    "NaiveCache",
)