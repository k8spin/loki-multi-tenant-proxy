# Syncing with upstream

## Assumptions

This guide uses ['Loki Distributed'](https://github.com/grafana/helm-charts/tree/main/charts/loki-distributed) chart as an example which is one of the charts held in upstream `helm-charts` repo. 

I want to easily track it, but also be able to submit patches to upstream as well as make some specific changes that
I never want to go to upstream.

## Repositories types and setup

We will be working with 3 git repositories per project. This covers the most complex scenario,
where we track upstream repo that hosts multiple charts in a single repository and we want to track
all of them, but finally get just a single chart from there.

Please note, that for a simpler scenario, where you just want to track a subdirectory of an upstream
repository and in general repository layout is very similar to what you have or when you don't plan to
submit patches to upstream, you can successfully use this method with just 2 repositories, "upstream"
and "chart repo".

The three repos we're going to use are:

- Original upstream repo <https://github.com/grafana/helm-charts/tree/main/charts/loki-distributed>.
  Every time I say "upstream" repo, I mean this one. It is read-only for us and maintaned by external
  organization.
- (Optional, but recommended) Our copy for tracking the "upstream". We will fork the upstream to
  create what we call "upstream copy"
  repo. This repo will be used for easily contributing changes into the upstream repo.
  In this repo we prepare patches we hope can be accepted by upstream. Our workflow allows us to submit
  a patch to "upstream" and either use the patch before it is accepted or wait until it is accepted
  upstream and shows in one of the branches of "upstream".
  If this repo is a multi-chart repo, where all the charts share some code, it is also a perfect place
  to apply patches to the shared code, whether we want to submit them to upstream or not.
- Chart repo. The "chart repo" is the one where we keep a single chart we want to build for the app
  platform. Here we apply all the changes and patches that we need to make it work, but also don't
  really want to send to "upstream" (so, any Giant Swarm specific stuff).
  
## Setting up repos

Everything here will be shown as an example based on the
[grafana's helm charts repo](https://github.com/grafana/helm-charts/). Please make
sure to go there and have a look at how the repo is organized before you read on.

### Upstream copy

Let's start with creating "upstream copy". Go on github to the "upstream" repo and fork it.
Make sure to change
the default repo name into something meaningful and ending with "-upstream". In my example, the
default repo name was `giantswarm/helm-charts`, but I changed it to
[`grafana-helm-charts-upstream`](https://github.com/giantswarm/grafana-helm-charts-upstream).

Now, clone the "upstream copy" repo to your machine:

```
git clone git@github.com:giantswarm/grafana-helm-charts-upstream.git
cd grafana-helm-charts-upstream
```

We will use the "upstream copy" repo in the following way:

- `main` (or `master`) - this is the branch we will use in "chart repo". It is tracking "upstream",
  but also will include
  all our patches to "upstream" that we hope to be accepted by upstream project some day, but need to
  use right away.
- `upstream-main` - to directly track `main`/`master` branch from "upstream". This branch is read-only
  for us; we only use it to synchronize with "upstream". We set it up by adding
  a new remote and setting merge config for it:
  
  ```
  git checkout -b "upstream-main"
  git remote add -f upstream https://github.com/grafana/helm-charts.git
  git branch -u upstream/main
  git push origin upstream-main
  # now we can pull changes from "upstream"/main and merge to our "upstream copy"/upstream-main  
  ```
- or using the new Git syntax:
  ```
  git remote add -f upstream https://github.com/grafana/helm-charts.git
  git switch -c upstream-main upstream/main # creates branch and sets tracking
  git push origin upstream-main
  ```

### Chart repo

Now, it's time to create our "chart repo" and reference the code we have in "upstream copy".
Go to `github` and create a new repo using the `ginatswarm-template-app` template. I've created
<https://github.com/giantswarm/loki-app>.

Clone this repo to your local machine and setup "upstream copy" as remote to track:

```
git clone git@github.com:giantswarm/loki-app.git
cd loki-app
git rm -r helm/{APP-NAME} && git commit -am "remove template chart" && git push  # optionally remove the chart template
git remote add -f --no-tags upstream-copy git@github.com:giantswarm/grafana-helm-charts-upstream.git  # add remote
```

Please not the `--no-tags` flag. If you add it, no tags from remote repo will be added to yours.
This means you won't "pollute" your local repo, which is a good thing, but also it will make impossible
to check what upstream means by specific tag. This might be useful, especially when you're migrating
to the subtree workflow. The decision to include tags is yours, but as a rule of thumb it's better to
not include them unless you know they are useful for you.

Now, we add code from "upstream-copy" as subtree. We have 2 options here:

1. We add the whole "upstream-copy" repo, as it is in `main` branch, as a subdirectory in the current
   repo
   
   ```
   git subtree add --prefix helm/loki-app upstream-copy main --squash
   git push
   ```
   
   That's it, now your `helm/loki-app` has the same content as is present in the `main` branch of
   `upstream-copy` remote. The `--squash` option squashes all the incoming commits into one big commit,
   which is a good thing, as otherwise you'll put all the commits from upstream into your local repo
   and make the history really noisy.
   
2. More complex scenario: we want to add only the `charts/loki-distributed` subdirectory from the
   `main` branch of `upstream-copy`. To do that, we first need to use the `subtree split` command to
   go over all commits and split only these that altered files in this directory into a temporary
   branch `temp-split-branch`. Then we add this branch as a subtree:

   ```
   git fetch upstream-copy main  # fetch the most recent state from "upstream copy/main"
   git checkout upstream-copy/main  # it's OK to be in detached head, we won't change anything
   git subtree split -P charts/helm-distributed -b temp-split-branch
   git checkout master
   git subtree add --squash -P helm/loki-distributed temp-split-branch
   ```
   
   Important: here we use `main` from `upstream-copy` as the state we want Most probably it makes more
   sense for you to use some other state of the `upstream-copy`, like a `vX.Y.Z` tag, which means a
   stable release of the chart. Here we're tracking the cutting edge in `main` repo.

## Workflows

### I want to setup my local repos after they were already created for the first time

- to setup "upstream-copy":

  ```
  git clone git@github.com:giantswarm/grafana-helm-charts-upstream.git
  cd grafana-helm-charts-upstream
  git checkout upstream-main
  git remote add -f upstream https://github.com/grafana/helm-charts.git
  ```

- to setup "chart repo"

  ```
  git clone git@github.com:giantswarm/loki-app.git
  cd loki-app
  git remote add -f --no-tags upstream-copy git@github.com:giantswarm/grafana-helm-charts-upstream.git  # add remote
  ```

### I want to update to the latest version from upstream

Assuming you want to get to the state of `main` branch in `upstream-main`. If you want any other state,
replace `upstream/main` with any other branch or just tag: `v.X.Y.Z` (to see upstream tags, you need to
skip the `--skip-tags` flag, as explained [above in set up instructions](#chart-repo)).

1. In "upstream-copy" repo
  - checkout "upstream-main" branch
  - fetch changes from "upstream/main", merge them with "upstream-main"
  - checkout "main", merge "upstream-main" to it, push "master"
  - example commands:

  ```
  git checkout upstream-main
  git fetch upstream
  git merge upstream/main  
  # create PR in github from upstream-main to main XOR run
  git checkout main
  git merge upstream-main
  git push
  ```

2. In "chart repo"
  - if the subtree is tracking the whole "upstream copy" repo
  
  ```
  git fetch upstream-copy main
  git subtree pull --prefix helm/loki-app upstream-copy main --squash  
  ```
  
  - if the subtree is tracking a subdir of "upstream copy":
  
  ```
  git fetch upstream-copy main  # fetch the most recent state from "upstream copy/main"
  git checkout upstream-copy/main  # it's OK to be in detached head, we won't change anything
  git subtree split -P charts/loki-distributed -b temp-split-branch
  git checkout master
  git subtree merge --squash -P helm/loki temp-split-branch
  git push 
  git branch -D temp-split-branch
  ```

### I want to send non-urgent patch for upstream

Do this if you want to submit a patch for "upstream", but you also want to wait it until it
is accepted by upstream (so, you'll get your patch applied and then get it from "upstream" someday):

- go to "upstream copy", update remote "upstream" and fetch changes into the "upstream-main" branch
- create a branch "my-feature" from "upstream-main"
- when ready, create a PR for "upstream"
- when PR is merged, remove local "my-feature" branch and update our dependencies as in [normal upstream update](#I-want-to-update-to-the-latest-version-from-upstream)
 
### I want to send urgent patch for upstream and use it already

Do this if you want to submit a patch for "upstream" and you need to use it right away, without waiting 
for being accepted by upstream:

- go to "upstream copy", update remote "upstream" and fetch changes into the "upstream-main" branch
- create a branch "my-feature" from "upstream-main"
- when ready, create a PR (PR1) for "upstream"
- create another PR (PR2) to merge "my-feature" into "main"
- when PR2 is merged, update "chart repo" dependency on "upstream-copy/master" as in point 2 in [normal upstream update](#I-want-to-update-to-the-latest-version-from-upstream)
- when PR1 is merged, remove local "my-feature" branch and update our dependencies as in [normal upstream update](#I-want-to-update-to-the-latest-version-from-upstream)

### I want to make changes that I don't want to be ever sent to upstream

Do this if you want to make any Giant Swarm specific changes to the chart. 
We have two options about where to do that and it's up to you to think where it makes the most
sense.

1. In the "chart repo" - this should be your default.

  - just do it - you can commit and change anything you want in the "subtree" catalog and your
    changes won't be lost when you update it.
  
2. In the "upstream copy" repo - makes sense for cases where multiple charts include some shared
   sub-chart and you want to patch it.

  - go to "upstream copy", checkout and update "main" branch
  - create a branch "my-feature" from "main"
  - when ready, create a PR from "my-feature" to "main"
  - when PR is merged, update "chart repo" dependency on "upstream-copy/master" as in point 2 in [normal upstream update](#I-want-to-update-to-the-latest-version-from-upstream)


### I want to switch from another way of tracking upstream to the git-subtree way

In general, we have two options here:

1) Git-supported. It works like this: we start by figuring our a commit (tag, branch, anything)
   in our current repo that was an exact copy of a known upstream version. Let's say this is
   reperesented by the `vX.Y.Z` tag. Now, I can save a diff between that clean state (a state of
   my repo when I got it from the "upstream" but before I applied any custom changes) and my
   current most recent state. The result should
   include everything we've changed since `vX.Y.Z` comparing to upstream.
   Now we remove the code from our repo, then include
   it again in the exactly same `vX.Y.Z` version, but this time using the `git-subtree` command.
   Then, we apply our
   patch file on the subtree and commit it. From now on, we can do any update as 
   [described above](#i-want-to-update-to-the-latest-version-from-upstream).
   
   Example: I want to switch my `grafana-app` repo to use git subtree from "upstream copy". I know
   that my repo in the commit `1111111` has the same code as the "upstream" repo had in
   the `grafana-6.1.3` tag. We also want to do the migration in separate branch `switch-to-subtree`
   to be able to create a valid PR for the change and not work on the `master` directly.
   
   ```
   git checkout -b switch-to-subtree
   git diff 1111111 -- helm/grafana-app > chart.diff
   git remove -r helm/grafana-app
   git commit -am "chart code deleted"
   git checkout grafana-6.1.3
   git subtree split -P charts/grafana -b temp-split-branch
   git checkout switch-to-subtree
   git subtree add --squash -P helm/grafana-app temp-split-branch # switched to subtree
   git apply chart.diff # applied any custom changes
   git commit -am "applied custom changes"
   # you're ready continue with updating to the current state
   ```
   
2) Manual way. We copy our current state somewhere (outside the current git tree), then we remove all
   the code we want to get from "upstream" or "upstream-copy". We add the code back using `git subtree`.
   Then we manually go over our backup copy and apply any chnages needed by editing the code. Then we
   commit the changes. From now on, we can do any update as 
   [described above](#i-want-to-update-to-the-latest-version-from-upstream).
   
   Example:
   
   ```
   git checkout -b switch-to-subtree
   cp -a helm/grafana-app /tmp
   git remove -r helm/grafana-app
   git commit -am "chart code deleted"
   git checkout grafana-7.1.0 # a version I want to update to
   git subtree split -P charts/grafana -b temp-split-branch
   git checkout switch-to-subtree
   git subtree add --squash -P helm/grafana-app temp-split-branch # switched to subtree
   # the hard part: compare what you have in /tmp/grafana-app with the helm/grafana-app and apply
   # all the missing changes
   git commit -am "applied custom changes"
   # you're ready continue with updating to the current state
   ```
