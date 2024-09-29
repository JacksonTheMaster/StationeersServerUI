##Work in the Private Repo:
Develop your code in the private dev repository, making commits as usual.

##Add the Public Repository as a Remote:
In your private repository, add the public repository as a remote by running:

'''bash
git remote add public-repo https://github.com/username/public-repo.git
Push Changes to Public Repository:
Once you’re ready to sync the private dev branch (or any feature branch) to the public nightly branch, push the changes from your private repository:
'''

'''bash
git push public-repo dev:nightly
This command will push the dev branch from your private repo to the nightly branch of the public repository.
'''
