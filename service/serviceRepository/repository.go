package serviceRepository

import (
  "MareWood/config"
  "MareWood/helper"
  "MareWood/models"
  "MareWood/sql"
  "errors"
  "os"
  "os/exec"
  "strconv"
  "strings"
)

func CloneRepo(repo * models.Repository, claims * models.Claims) {
  out, err: = GitClone(strconv.Itoa(int(repo.ID)), repo.Url, repo.UserName, repo.Password)
  if err != nil {
    sql.DB.Model(repo).
    Where("id = ?", repo.ID).
    Update("status", models.RepoStatusFail).
    Update("terminal_info", out)
    return
  }

  sql.DB.Model(repo).
  Where("id = ?", repo.ID).
  Update("status", models.RepoStatusSuccess).
  Update("terminal_info", out)

  models.Broadcast < -models.Message {
    Type: models.MsgTypeSuccess,
    TriggerID: claims.ID,
    TriggerUsername: claims.Username,
    NeedNotifySelf: true,
    Message: claims.Username + "'s Repository, " + repo.Name + " has been successfully cloned",
  }
}

//克隆仓库，userName，password可留空
func GitClone(repositoryId string, gitUrl string, userName string, password string)(string, error) {

  var (
    cmd * exec.Cmd endingUrl string
  )
  endingUrl = gitUrl
  if userName != "" && password != "" {
    authUrl, err: = helper.GitUrl2AuthUrl(gitUrl, userName, password)
    if err != nil {
      return "", err
    }
    endingUrl = authUrl
  }
  cmd = exec.Command("git", "clone", endingUrl)
  cmd.Dir = config.Cfg.RepositoryDir
  out, err: = cmd.CombinedOutput()
  if err != nil {
    return string(out), err
  }
  repositoryName, err: = helper.GetRepositoryNameByUrl(gitUrl)

  if err != nil {
    return "", err
  }

  repositoryDir: = config.Cfg.RepositoryDir + "/" + repositoryName
  newRepositoryDir: = config.Cfg.RepositoryDir + "/" + repositoryId

  err = os.Rename(repositoryDir, newRepositoryDir)
  if err != nil {
    return "", err
  }

  return string(out), nil
}

func GitPull(repositoryId string)(string, error) {
  return RunCmdOnRepositoryDir(repositoryId, "git", "pull")
}

func DiscardChange(repositoryId string)(string, error) {
  return RunCmdOnRepositoryDir(repositoryId, "git", "checkout", ".")
}

func DeleteRepository(repositoryId string) error {

  repoDir: = config.Cfg.RepositoryDir + "/" + repositoryId

    if !helper.IsDir(repoDir) {
    return nil
  }

  return os.RemoveAll(repoDir)
}

func PruneBranch(repositoryId string)(string, error) {
  if out, err: = GitCheckout(repositoryId, "master");
  err != nil {
    return out, err
  }
  if out, err: = GitPull(repositoryId);
  err != nil {
      return out, err
    }
    //裁剪分支
  return RunCmdOnRepositoryDir(repositoryId, "git", "remote", "prune", "origin")
}

func GetBranch(repositoryId string)([] string, error) {

  out, err: = RunCmdOnRepositoryDir(repositoryId, "git", "branch", "-r")

  if err != nil {
    return [] string {}, err
  }

  deleteOrigin: = strings.ReplaceAll(string(out), "origin/", "")

  branch: = strings.Split(strings.Trim(strings.ReplaceAll(deleteOrigin, " ", ""), "\n"), "\n")

  var newBranch[] string
  for _, b: = range branch {
    if !strings.Contains(b, "HEAD") {
      newBranch = append(newBranch, b)
    }
  }
  return newBranch, nil
}

func GitCheckout(repositoryId string, branch string)(string, error) {

  return RunCmdOnRepositoryDir(repositoryId, "git", "checkout", branch)
}

// ADD 优先用bun 失败再用 npm
type ExceptionStruct struct {
  Try func()
  Catch func(Exception)
  Finally func()
}
type Exception interface {}
func Throw(up Exception) {
  panic(up)
}
func(this ExceptionStruct) Do() {
  if this.Finally != nil {

    defer this.Finally()
  }
  if this.Catch != nil {
    defer func() {
      if e: = recover();
      e != nil {
        this.Catch(e)
      }
    }()
  }
  this.Try()
}


//仓库URL， 构建命令 test、build、build:dev
func RunBuild(repositoryId string, buildCmd string)(string, error) {
  ExceptionStruct {
    Try: func() {
      return RunCmdOnRepositoryDir(repositoryId, "bun", "run", buildCmd)
    },
    Catch: func(e Exception) {
      fmt.Printf("exception %v\n", e)
      return RunCmdOnRepositoryDir(repositoryId, "npm", "run", buildCmd)
    }
  }.Do()

}

func RunCmdOnRepositoryDir(repositoryId string, cmdName string, arg...string)(string, error) {

  repositoryDir: = config.Cfg.RepositoryDir + "/" + repositoryId

    if !helper.IsDir(repositoryDir) {
    return "", errors.New("Can't find repository dir=>" + repositoryDir)
  }

  cmd: = exec.Command(cmdName, arg...)
  cmd.Dir = repositoryDir

  out,
  err: = cmd.CombinedOutput()
  if err != nil {
    return "", errors.New(cmdName + "Command exited unexpectedly=>\n" + string(out))
  }
  return string(out),
  nil
}