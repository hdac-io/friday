class DeadDaemonException(Exception):
    def __str__(self):
        return "already dead daemon"

class FinishedWithError(Exception):
    def __str__(self):
        return "process was finished with error"
