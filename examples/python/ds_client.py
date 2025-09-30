#!/usr/bin/env python3
"""
ds-go Python Client
A simple client library for interacting with the ds-go API
"""

import requests
import json
from typing import Dict, List, Optional, Any
from datetime import datetime


class DsClient:
    """Client for ds-go repository manager API"""

    def __init__(self, base_url: str = "http://127.0.0.1:7777/v1"):
        """Initialize the client"""
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()

    def _request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        """Make an API request"""
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, **kwargs)
        response.raise_for_status()
        return response.json()

    def capabilities(self) -> Dict[str, Any]:
        """Get API capabilities and metadata"""
        return self._request("GET", "/capabilities")

    def status(self,
               dirty: Optional[bool] = None,
               account: Optional[str] = None,
               path: Optional[str] = None) -> Dict[str, Any]:
        """Get repository status"""
        params = {}
        if dirty is not None:
            params["dirty"] = str(dirty).lower()
        if account:
            params["account"] = account
        if path:
            params["path"] = path
        return self._request("GET", "/status", params=params)

    def scan(self, path: Optional[str] = None) -> Dict[str, Any]:
        """Scan for repositories"""
        params = {"path": path} if path else {}
        return self._request("GET", "/scan", params=params)

    def fetch(self,
              account: Optional[str] = None,
              dirty: Optional[bool] = None) -> Dict[str, Any]:
        """Fetch remote information"""
        params = {}
        if account:
            params["account"] = account
        if dirty is not None:
            params["dirty"] = str(dirty).lower()
        return self._request("GET", "/fetch", params=params)

    def organize_plan(self, require_clean: bool = False) -> Dict[str, Any]:
        """Get repository organization plan"""
        params = {"require_clean": str(require_clean).lower()}
        return self._request("GET", "/organize/plan", params=params)

    def organize_apply(self,
                       require_clean: bool = False,
                       force: bool = False,
                       dry_run: bool = False) -> Dict[str, Any]:
        """Apply repository organization"""
        params = {
            "require_clean": str(require_clean).lower(),
            "force": str(force).lower(),
            "dry_run": str(dry_run).lower()
        }
        return self._request("POST", "/organize/apply", params=params)

    def policy_check(self,
                     file: str = ".project-compliance.yaml",
                     fail_on: str = "critical") -> Dict[str, Any]:
        """Run policy compliance checks"""
        params = {"file": file, "fail_on": fail_on}
        return self._request("GET", "/policy/check", params=params)

    def exec_command(self,
                     cmd: str,
                     account: Optional[str] = None,
                     dirty: Optional[bool] = None,
                     timeout: int = 30) -> Dict[str, Any]:
        """Execute command across repositories"""
        params = {"timeout": timeout}
        if account:
            params["account"] = account
        if dirty is not None:
            params["dirty"] = str(dirty).lower()

        json_data = {"cmd": cmd}
        return self._request("POST", "/exec", params=params, json=json_data)

    def get_dirty_repos(self) -> List[Dict[str, Any]]:
        """Get list of repositories with uncommitted changes"""
        response = self.status(dirty=True)
        return response["data"]["repositories"]

    def get_repos_behind(self) -> List[Dict[str, Any]]:
        """Get repositories that are behind remote"""
        response = self.status()
        repos = response["data"]["repositories"]
        return [r for r in repos if r.get("behind", 0) > 0]

    def get_summary(self) -> Dict[str, Any]:
        """Get repository summary statistics"""
        response = self.status()
        return response["data"]["summary"]


def main():
    """Example usage"""
    client = DsClient()

    # Check capabilities
    print("API Capabilities:")
    caps = client.capabilities()
    print(f"  Version: {caps['data']['version']}")
    print(f"  Features: {caps['data']['features']}")
    print()

    # Get repository summary
    print("Repository Summary:")
    summary = client.get_summary()
    print(f"  Total: {summary['total']}")
    print(f"  Clean: {summary['clean']}")
    print(f"  Dirty: {summary['dirty']}")
    print(f"  Ahead: {summary['ahead']}")
    print(f"  Behind: {summary['behind']}")
    print()

    # Get dirty repositories
    dirty_repos = client.get_dirty_repos()
    if dirty_repos:
        print(f"Dirty Repositories ({len(dirty_repos)}):")
        for repo in dirty_repos[:5]:  # Show first 5
            print(f"  - {repo['name']}: {repo['uncommitted_files']} files")
        if len(dirty_repos) > 5:
            print(f"  ... and {len(dirty_repos) - 5} more")
        print()

    # Check repos behind remote
    behind_repos = client.get_repos_behind()
    if behind_repos:
        print(f"Repositories Behind Remote ({len(behind_repos)}):")
        for repo in behind_repos[:5]:
            print(f"  - {repo['name']}: {repo['behind']} commits behind")
        if len(behind_repos) > 5:
            print(f"  ... and {len(behind_repos) - 5} more")
        print()

    # Check policy compliance
    print("Policy Compliance Check:")
    try:
        policy = client.policy_check()
        policy_summary = policy["data"]["summary"]
        print(f"  Total Checks: {policy_summary['total']}")
        print(f"  Passed: {policy_summary['passed']}")
        print(f"  Failed: {policy_summary['failed']}")
        print(f"  Threshold Failed: {policy['data']['failed_threshold']}")
    except Exception as e:
        print(f"  Policy check failed: {e}")


if __name__ == "__main__":
    main()