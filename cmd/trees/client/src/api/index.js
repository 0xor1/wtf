import axios from 'axios'

let memCache = {}

let newApi = (isMDoApi) => {
  let mDoSending = false
  let mDoSent = false
  let awaitingMDoList = []
  let doReq = (path, args, headers) => {
    path = '/api'+path
    if (!isMDoApi || (isMDoApi && mDoSending && !mDoSent)) {
      headers = headers || {"X-Client": "tlbx-web-client"}
      return axios({
        method: 'put',
        url: path,
        headers: headers,
        data: args
      }).then((res) => {
        return res.data
      }).catch((err) => {
        throw {
          status: err.response.status,
          body: err.response.data
        }
      })
    } else if (isMDoApi && !mDoSending && !mDoSent) {
      let awaitingMDoObj = {
        path: path,
        args: args,
        resolve: null,
        reject: null
      }
      awaitingMDoList.push(awaitingMDoObj)
      return new Promise((resolve, reject) => {
        awaitingMDoObj.resolve = resolve
        awaitingMDoObj.reject = reject
      })
    } else {
      throw new Error('invalid get call, use the default api object or a new mdo instance from api.newMDoApi()')
    }
  }

  return {
    newMDoApi: () => {
      return newApi(true)
    },
    sendMDo: () => {
      if (!isMDoApi) {
        throw new Error('MDoes must be made from the api instance returned from api.newMDoApi()')
      } else if (mDoSending || mDoSent) {
        throw new Error('each MDo must be started with a fresh api.newMDoApi(), once used that same instance cannot be reused')
      }
      mDoSending = true
      let asyncIndividualPromisesReady
      asyncIndividualPromisesReady = (resolve) => {
        let ready = true
        for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
          if (awaitingMDoList[i].resolve === null) {
            ready = false
            setTimeout(asyncIndividualPromisesReady, 0, resolve)
          }
        }
        if (ready) {
          resolve()
        }
      }
      let mdoErrors = []
      mdoErrors.isMDoErrors = true
      let mDoComplete = false
      let mDoCompleterFunc
      mDoCompleterFunc = (resolve, reject) => {
        if (mDoComplete) {
          if (mdoErrors.length === 0) {
            resolve()
          } else {
            reject(mdoErrors)
          }
        } else {
          setTimeout(mDoCompleterFunc, 0, resolve, reject)
        }
      }
      new Promise(asyncIndividualPromisesReady).then(() => {
        let mDoObj = {}
        for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
          let key = '' + i
          mDoObj[key] = {
            path: awaitingMDoList[i].path,
            args: awaitingMDoList[i].args
          }
        }
        doReq('/mdo', mDoObj).then((res) => {
          for (let i = 0, l = awaitingMDoList.length; i < l; i++) {
            let key = '' + i
            if (res[key].status === 200) {
              awaitingMDoList[i].resolve(res[key].body)
            } else {
              mdoErrors.push(res[key])
              awaitingMDoList[i].reject(res[key])
            }
          }
        }).catch((error) => {
          mdoErrors.push(error)
          for (let i = 0, il = awaitingMDoList.length; i < il; i++) {
            awaitingMDoList[i].reject(error)
          }
        }).finally(()=>{
          mDoComplete = true
          mDoSending = false
          mDoSent = true
        })
      })
      return new Promise(mDoCompleterFunc)
    },
    user: {
      register: (alias, handle, email, pwd, confirmPwd) => {
        return doReq('/user/register', {alias, handle, email, pwd, confirmPwd})
      },
      resendActivateLink: (email) => {
        return doReq('/user/resendActivateLink', {email})
      },
      activate: (email, code) => {
        return doReq('/user/activate', {email, code})
      },
      changeEmail: (newEmail) => {
        return doReq('/user/changeEmail', {newEmail})
      },
      resendChangeEmailLink: () => {
        return doReq('/user/resendChangeEmailLink')
      },
      confirmChangeEmail: (me, code) => {
        return doReq('/user/confirmChangeEmail', {me, code})
      },
      resetPwd: (email) => {
        return doReq('/user/resetPwd', {email})
      },
      setHandle: (handle) => {
        return doReq('/user/setHandle', {handle: handle}).then(()=>{
          memCache.me.handle = handle
        })
      },
      setAlias: (alias) => {
        return doReq('/user/setAlias', {alias}).then(()=>{
          memCache.me.alias = alias
        })
      },
      setAvatar: (avatar) => {
        return doReq('/user/setAvatar', avatar).then(()=>{
          memCache.me.hasAvatar = avatar === null
        })
      },
      setPwd: (currentPwd, newPwd, confirmNewPwd) => {
        return doReq('/user/setPwd', {currentPwd, newPwd, confirmNewPwd})
      },
      delete: (pwd) => {
        return doReq('/user/delete', {pwd})
      },
      login: (email, pwd) => {
        return doReq('/user/login', {email, pwd}).then((res)=>{
          memCache.me = res
          memCache[res.id] = res
          return res
        })
      },
      logout: () => {
        memCache = {}
        return doReq('/user/logout')
      },
      me: () => {
        if (memCache.me) {
          return new Promise((resolve) => {
            resolve(memCache.me)
          })
        }
        return doReq('/user/me').then((res) => {
          memCache.me = res
          memCache[res.id] = res
          return res
        })
      },
      get: (ids) => {
        let toGet = []
        let found = []
        ids.forEach((id)=>{
          if (memCache[id]) {
            found.push(memCache[id])
          } else {
            found.push(memCache[id])
            toGet.push(id)
          }
        })
        if (toGet.length === 0) {
          return new Promise((resolve) => {
            resolve(found)
          })
        }
        return doReq('/user/get', {users: toGet}).then((res) => {
          if (res != null) {
            res.forEach((user)=>{
              memCache[user.id] = user
            })
          }
          return res
        })
      }
    },
    project: {
      create: (currencyCode, hoursPerDay, daysPerWeek, startOn, dueOn, isPublic, name) => {
        return doReq('/project/create', {currencyCode, hoursPerDay, daysPerWeek, startOn, dueOn, isPublic, name})
      },
      get: (host, ids, namePrefix, isArchived, isPublic, createdOnMin, createdOnMax, startOnMin, startOnMax, dueOnMin, dueOnMax, after, sort, asc, limit) => {
        return doReq('/project/get', {host, ids, namePrefix, isArchived, isPublic, createdOnMin, createdOnMax, startOnMin, startOnMax, dueOnMin, dueOnMax, after, sort, asc, limit})
      },
      update: (id, name, currencyCode, hoursPerDay, daysPerWeek, startOn, dueOn, isArchived, isPublic) => {
        return doReq('/project/update', [{id, name, currencyCode, hoursPerDay, daysPerWeek, startOn, dueOn, isArchived, isPublic}])
      },
      delete: (ids) => {
        return doReq('/project/delete', ids)
      },
      addUsers: (host, project, users) => {
        return doReq('/project/addUsers', {host, project, users})
      },
      getMe: (host, project) => {
        return doReq('/project/getMe', {host, project})
      },
      getUsers: (host, project, ids, role, handlePrefix, after, limit) => {
        return doReq('/project/getUsers', {host, project, ids, role, handlePrefix, after, limit})
      },
      setUserRoles: (host, project, users) => {
        return doReq('/project/setUserRoles', {host, project, users})
      },
      removeUsers: (host, project, users) => {
        return doReq('/project/removeUsers', {host, project, users})
      },
      getActivities: (host, project, task, item, user, occuredAfter, occuredBefore, limit) => {
        return doReq('/project/getActivities', {host, project, task, item, user, occuredAfter, occuredBefore, limit})
      }
    },
    task: {
      create: (host, project, parent, previousSibling, name, description, isParallel, user, estimatedTime, estimatedExpense) => {
        return doReq('/task/create', {host, project, parent, previousSibling, name, description, isParallel, user, estimatedTime, estimatedExpense})
      },
      update: (host, project, id, parent, previousSibling, name, description, isParallel, user, estimatedTime, estimatedExpense) => {
        return doReq('/task/update', {host, project, id, parent, previousSibling, name, description, isParallel, user, estimatedTime, estimatedExpense})
      },
      delete: (host, project, id) => {
        return doReq('/task/delete', {host, project, id})
      },
      get: (host, project, id) => {
        return doReq('/task/get', {host, project, id})
      },
      getAncestors: (host, project, id, limit) => {
        return doReq('/task/getAncestors', {host, project, id, limit})
      },
      getChildren: (host, project, id, after, limit) => {
        return doReq('/task/getChildren', {host, project, id, after, limit})
      }
    },
    time: {
      create: (host, project, task, duration, note) => {
        return doReq('/time/create', {host, project, task, duration, note})
      },
      update: (host, project, task, id, duration, note) => {
        return doReq('/time/update', {host, project, task, id, duration, note})
      },
      get: (host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit) => {
        return doReq('/time/get', {host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit})
      },
      delete: (host, project, task, id) => {
        return doReq('/time/delete', {host, project, task, id})
      }
    },
    expense: {
      create: (host, project, task, value, note) => {
        return doReq('/expense/create', {host, project, task, value, note})
      },
      update: (host, project, task, id, value, note) => {
        return doReq('/expense/update', {host, project, task, id, value, note})
      },
      get: (host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit) => {
        return doReq('/expense/get', {host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit})
      },
      delete: (host, project, task, id) => {
        return doReq('/expense/delete', {host, project, task, id})
      }
    },
    file: {
      create: (host, project, task, name, mimeType, size, content) => {
        return doReq('/file/getPresignedPutUrl', {host, project, task, name, mimeType, size}).then((res)=>{
          let id = res.id
          doReq(res.url, content, {
            "Host": (new URL(res.url)).hostname,
            "X-Amz-Acl": "private",
            "Content-Length": size, 
            "Content-Type": mimeType,
            "Content-Disposition": "attachment; filename="+name,
          }).then(()=>{
            doReq("/file/finalize", {host, project, task, id})
          })
        })
      },
      getPresignedGetUrl: (host, project, task, id, isDownload) => {
        return doReq('/file/getPresignedGetUrl', {host, project, task, id, isDownload})
      },
      get: (host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit) => {
        return doReq('/file/get', {host, project, task, ids, createOnMin, createdOnMax, createdBy, after, asc, limit})
      },
      delete: (host, project, task, id) => {
        return doReq('/file/delete', {host, project, task, id})
      }
    },
    comment: {
      create: (host, project, task, body) => {
        return doReq('/comment/create', {host, project, task, body})
      },
      update: (host, project, task, id, body) => {
        return doReq('/comment/update', {host, project, task, id, body})
      },
      get: (host, project, task, after, limit) => {
        return doReq('/comment/get', {host, project, task, after, limit})
      },
      delete: (host, project, task, id) => {
        return doReq('/comment/delete', {host, project, task, id})
      }
    }
  }
}

// make it available for console hacking
window.api = newApi(false)
export default window.api