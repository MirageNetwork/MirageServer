package controller

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

type TagsData struct {
	TagOwners []Tag `json:"tagOwners"`
}
type Tag struct {
	TagName string   `json:"tagName"`
	Owners  []string `json:"owners"`
}

// 接受/admin/api/acls/tags的Get请求，用于查询tags
func (h *Mirage) CAPIGetTags(
	w http.ResponseWriter,
	r *http.Request,
) {
	user := h.verifyTokenIDandGetUser(w, r)
	if user.CheckEmpty() {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}

	resData := TagsData{
		TagOwners: []Tag{},
	}
	org, err := h.GetOrgnaizationByID(user.OrganizationID)
	if err != nil {
		h.doAPIResponse(w, "用户组织信息获取失败", nil)
	}
	for tagName, owners := range org.AclPolicy.TagOwners {
		resData.TagOwners = append(resData.TagOwners, Tag{
			TagName: tagName,
			Owners:  owners,
		})
	}

	h.doAPIResponse(w, "", resData)
}

// 请求报文：
type CreateTagREQ struct {
	State   string   `json:"state"`
	TagName string   `json:"tagName"`
	Owners  []string `json:"owners"`
}

// 接受/admin/api/acls/tags的Post请求，用于创建标签
func (h *Mirage) CAPIPostTags(
	w http.ResponseWriter,
	r *http.Request,
) {
	user := h.verifyTokenIDandGetUser(w, r)
	if user.CheckEmpty() {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}
	err := r.ParseForm()
	if err != nil {
		h.doAPIResponse(w, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := CreateTagREQ{}
	json.NewDecoder(r.Body).Decode(&reqData)
	switch reqData.State {
	case "create":
		org, err := h.GetOrgnaizationByID(user.OrganizationID)
		if err != nil {
			h.doAPIResponse(w, "用户组织信息获取失败", nil)
		}
		acl := org.AclPolicy
		if _, ok := acl.TagOwners["tag:"+reqData.TagName]; ok {
			h.doAPIResponse(w, "occupied", nil)
			return
		}
		acl.TagOwners["tag:"+reqData.TagName] = reqData.Owners
		//aclPath := AbsolutePathFromConfigPath(ACLPath)
		//err = h.SaveACLPolicy(aclPath)
		err = h.SaveACLPolicyOfOrg(org)
		if err != nil {
			//delete(acl.TagOwners, "tag:"+reqData.TagName)
			h.doAPIResponse(w, "保存ACL策略失败:"+err.Error(), nil)
			return
		}
		/*
			err = h.UpdateACLRulesOfOrg(org)
			h.organizationCache.Delete(org.Name)
			if err != nil {
				delete(h.aclPolicy.TagOwners, "tag:"+reqData.TagName)
				h.doAPIResponse(w, "更新ACL规则失败:"+err.Error(), nil)
				return
			}
		*/
		resData := CreateTagREQ{
			TagName: reqData.TagName,
			Owners:  reqData.Owners,
		}
		h.doAPIResponse(w, "", resData)
	}
}

// 注销Key执行DELETE方法api/acls/tags/:tag
func (h *Mirage) CAPIDelTags(
	w http.ResponseWriter,
	r *http.Request,
) {
	user := h.verifyTokenIDandGetUser(w, r)
	if user.CheckEmpty() {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}
	org, err := h.GetOrgnaizationByID(user.OrganizationID)
	if err != nil {
		h.doAPIResponse(w, "用户组织信息获取失败", nil)
	}
	targetTagName := strings.TrimPrefix(r.URL.Path, "/admin/api/acls/tags/")
	_, ok := org.AclPolicy.TagOwners["tag:"+targetTagName]
	if !ok {
		h.doAPIResponse(w, "该标签不存在", nil)
		return
	}
	delete(org.AclPolicy.TagOwners, "tag:"+targetTagName)
	err = h.SaveACLPolicyOfOrg(org)
	if err != nil {
		h.doAPIResponse(w, "保存ACL策略失败:"+err.Error(), nil)
		return
	}
	/*
		delete(h.aclPolicy.TagOwners, "tag:"+targetTagName)
		aclPath := AbsolutePathFromConfigPath(ACLPath)
		err = h.SaveACLPolicy(aclPath)
		if err != nil {
			h.aclPolicy.TagOwners["tag:"+targetTagName] = owners
			h.doAPIResponse(w, "保存ACL策略失败:"+err.Error(), nil)
			return
		}
		err = h.UpdateACLRules()
		if err != nil {
			h.aclPolicy.TagOwners["tag:"+targetTagName] = owners
			h.doAPIResponse(w, "更新ACL规则失败:"+err.Error(), nil)
			return
		}
	*/
	h.doAPIResponse(w, "", targetTagName)
}
