import os
from OpenAI
from tenacity import retry, stop_after_attempt, wait_exponential

MAX_DIFF_SIZE = 12000  # prevent token overflow

client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))


def load_diff():
    try:
        with open("diff.txt", "r") as f:
            diff = f.read()
            return diff[:MAX_DIFF_SIZE]
    except Exception as e:
        return f"Error loading diff: {str(e)}"


@retry(stop=stop_after_attempt(3), wait=wait_exponential(min=2, max=10))
def call_ai(diff):
    prompt = f"""
You are a senior Python code reviewer.

Analyze the following code changes and provide:

1. Summary
