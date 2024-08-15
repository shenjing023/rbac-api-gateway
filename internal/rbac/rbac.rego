package rbac

import future.keywords.if
import future.keywords.in

default allow = false

# 允许管理员执行所有操作
allow if {
    input.user.role == "admin"
}

# 允许版主管理所有资源
allow if {
    input.user.role == "moderator"
    input.action in ["POST:/posts", "GET:/posts", "GET:/posts/:id", "PUT:/posts/:id", "DELETE:/posts/:id"]
}

# 允许普通用户执行基本操作
allow if {
    input.user.role == "user"
    input.action in ["POST:/posts", "GET:/posts", "GET:/posts/:id"]
}

# 允许用户更新或删除自己的资源
allow if {
    input.user.role == "user"
    input.action in ["PUT:/posts/:id", "DELETE:/posts/:id"]
    input.resource.is_owner == true
}