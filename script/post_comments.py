import os
import requests

repo = os.getenv("GITHUB_REPOSITORY")
token = os.getenv("GITHUB_TOKEN")
pr_number = os.getenv("GITHUB_REF").split("/")[-1]

headers = {
    "Authorization": f"Bearer {token}",
    "Accept": "application/vnd.github+json",
}


def get_existing_comments():
    url = f"https://api.github.com/repos/{repo}/issues/{pr_number}/comments"
    r = requests.get(url, headers=headers)
    return r.json() if r.status_code == 200 else []


def already_posted(comments):
    for c in comments:
        if "🤖 AI Python Code Review" in c.get("body", ""):
            return True
    return False


def post_comment(body):
    url = f"https://api.github.com/repos/{repo}/issues/{pr_number}/comments"
    requests.post(url, headers=headers, json={"body": body})


def main():
    try:
        with open("review.txt", "r") as f:
            review = f.read()
    except Exception:
        review = "⚠️ No review output generated."

    comments = get_existing_comments()

    if already_posted(comments):
        print("Comment already exists. Skipping.")
        return

    body = f"""## 🤖 AI Python Code Review

{review}

---
_This review was generated automatically. Please validate suggestions before applying._
"""

    post_comment(body)


if __name__ == "__main__":
    main()