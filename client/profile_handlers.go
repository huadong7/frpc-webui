package client

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/fatedier/frp/client/http/model"
	"github.com/fatedier/frp/client/proxy"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	httppkg "github.com/fatedier/frp/pkg/util/http"
	"github.com/fatedier/frp/pkg/util/jsonx"
	netpkg "github.com/fatedier/frp/pkg/util/net"
)

func registerProfileRoutes(helper *httppkg.RouterRegisterHelper, m *FrpcManager) {
	// Health endpoint without auth
	helper.Router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API routes with auth
	api := helper.Router.PathPrefix("/api").Subrouter()
	api.Use(helper.AuthMiddleware)
	api.Use(httppkg.NewRequestLogger)

	// Profile CRUD
	api.HandleFunc("/profiles", httppkg.MakeHTTPHandlerFunc(handleListProfiles(m))).Methods(http.MethodGet)
	api.HandleFunc("/profiles", httppkg.MakeHTTPHandlerFunc(handleCreateProfile(m))).Methods(http.MethodPost)
	api.HandleFunc("/profiles/{name}", httppkg.MakeHTTPHandlerFunc(handleGetProfile(m))).Methods(http.MethodGet)
	api.HandleFunc("/profiles/{name}", httppkg.MakeHTTPHandlerFunc(handleUpdateProfile(m))).Methods(http.MethodPut)
	api.HandleFunc("/profiles/{name}", httppkg.MakeHTTPHandlerFunc(handleDeleteProfile(m))).Methods(http.MethodDelete)
	api.HandleFunc("/profiles/{name}/start", httppkg.MakeHTTPHandlerFunc(handleStartProfile(m))).Methods(http.MethodPost)
	api.HandleFunc("/profiles/{name}/stop", httppkg.MakeHTTPHandlerFunc(handleStopProfile(m))).Methods(http.MethodPost)

	// Aggregate proxy status
	api.HandleFunc("/status", httppkg.MakeHTTPHandlerFunc(handleStatus(m))).Methods(http.MethodGet)

	// Backward compat: old config file endpoints (return empty in manager mode)
	api.HandleFunc("/config", httppkg.MakeHTTPHandlerFunc(handleGetConfig(m))).Methods(http.MethodGet)
	api.HandleFunc("/config", httppkg.MakeHTTPHandlerFunc(handlePutConfig(m))).Methods(http.MethodPut)
	api.HandleFunc("/reload", httppkg.MakeHTTPHandlerFunc(handleReload(m))).Methods(http.MethodGet)

	// Per-profile running status
	api.HandleFunc("/profiles/{name}/status", httppkg.MakeHTTPHandlerFunc(handleProfileStatus(m))).Methods(http.MethodGet)

	// Profile-scoped proxy CRUD
	api.HandleFunc("/profiles/{name}/proxies", httppkg.MakeHTTPHandlerFunc(handleListProfileProxies(m))).Methods(http.MethodGet)
	api.HandleFunc("/profiles/{name}/proxies", httppkg.MakeHTTPHandlerFunc(handleCreateProfileProxy(m))).Methods(http.MethodPost)
	api.HandleFunc("/profiles/{name}/proxies/{proxyName}", httppkg.MakeHTTPHandlerFunc(handleGetProfileProxy(m))).Methods(http.MethodGet)
	api.HandleFunc("/profiles/{name}/proxies/{proxyName}", httppkg.MakeHTTPHandlerFunc(handleUpdateProfileProxy(m))).Methods(http.MethodPut)
	api.HandleFunc("/profiles/{name}/proxies/{proxyName}", httppkg.MakeHTTPHandlerFunc(handleDeleteProfileProxy(m))).Methods(http.MethodDelete)

	// Profile-scoped visitor CRUD
	api.HandleFunc("/profiles/{name}/visitors", httppkg.MakeHTTPHandlerFunc(handleListProfileVisitors(m))).Methods(http.MethodGet)
	api.HandleFunc("/profiles/{name}/visitors", httppkg.MakeHTTPHandlerFunc(handleCreateProfileVisitor(m))).Methods(http.MethodPost)
	api.HandleFunc("/profiles/{name}/visitors/{visitorName}", httppkg.MakeHTTPHandlerFunc(handleGetProfileVisitor(m))).Methods(http.MethodGet)
	api.HandleFunc("/profiles/{name}/visitors/{visitorName}", httppkg.MakeHTTPHandlerFunc(handleUpdateProfileVisitor(m))).Methods(http.MethodPut)
	api.HandleFunc("/profiles/{name}/visitors/{visitorName}", httppkg.MakeHTTPHandlerFunc(handleDeleteProfileVisitor(m))).Methods(http.MethodDelete)

	// Profile-scoped frps dashboard query
	api.HandleFunc("/profiles/{name}/ports/used", httppkg.MakeHTTPHandlerFunc(handleGetProfileUsedPorts(m))).Methods(http.MethodGet)

	// Static files - registered on main router (not /api), with auth
	subRouter := helper.Router.NewRoute().Subrouter()
	subRouter.Use(helper.AuthMiddleware)
	subRouter.Handle("/favicon.ico", http.FileServer(helper.AssetsFS)).Methods("GET")
	subRouter.PathPrefix("/static/").Handler(
		netpkg.MakeHTTPGzipHandler(http.StripPrefix("/static/", http.FileServer(helper.AssetsFS))),
	).Methods("GET")
	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
}

// ─── Profile Handlers ──────────────────────────────────

func handleListProfiles(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		return map[string]any{"profiles": m.ListProfiles()}, nil
	}
}

func handleCreateProfile(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		body, err := ctx.Body()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "read body error: "+err.Error())
		}
		var entry ProfileEntry
		if err := jsonx.Unmarshal(body, &entry); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "parse JSON error: "+err.Error())
		}
		if entry.Config.Name == "" {
			return nil, httppkg.NewError(http.StatusBadRequest, "profile name is required")
		}
		created, err := m.CreateProfile(entry)
		if err != nil {
			return nil, httppkg.NewError(http.StatusConflict, err.Error())
		}
		return created, nil
	}
}

func handleGetProfile(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		entry, status, err := m.GetProfile(name)
		if err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		return map[string]any{
			"config": entry.Config, "status": status.Status,
			"runID": status.RunID, "error": status.Error,
			"proxies": entry.Proxies, "visitors": entry.Visitors,
		}, nil
	}
}

func handleUpdateProfile(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		body, err := ctx.Body()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "read body error: "+err.Error())
		}
		var entry ProfileEntry
		if err := jsonx.Unmarshal(body, &entry); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "parse JSON error: "+err.Error())
		}
		if entry.Config.Name != "" && entry.Config.Name != name {
			return nil, httppkg.NewError(http.StatusBadRequest,
				fmt.Sprintf("name in body %q does not match URL %q", entry.Config.Name, name))
		}
		entry.Config.Name = name
		updated, err := m.UpdateProfile(name, entry)
		if err != nil {
			return nil, httppkg.NewError(http.StatusConflict, err.Error())
		}
		return updated, nil
	}
}

func handleDeleteProfile(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		if err := m.DeleteProfile(name); err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		return nil, nil
	}
}

func handleStartProfile(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		if err := m.StartProfile(name); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		return nil, nil
	}
}

func handleStopProfile(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		if err := m.StopProfile(name); err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		return nil, nil
	}
}

// ─── Status Handler ────────────────────────────────────

func handleStatus(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		allStatuses := m.GetAllProxyStatus()
		res := make(model.StatusResp)
		for profileName, statuses := range allStatuses {
			serverAddr := m.GetProfileServerAddr(profileName)
			for _, status := range statuses {
				res[profileName+"/"+status.Type] = append(
					res[profileName+"/"+status.Type],
					buildProxyStatusResp(status, serverAddr),
				)
			}
		}
		return res, nil
	}
}

func buildProxyStatusResp(s *proxy.WorkingStatus, serverAddr string) model.ProxyStatusResp {
	psr := model.ProxyStatusResp{
		Name: s.Name, Type: s.Type, Status: s.Phase, Err: s.Err,
	}
	baseCfg := s.Cfg.GetBaseConfig()
	if baseCfg.LocalPort != 0 {
		psr.LocalAddr = net.JoinHostPort(baseCfg.LocalIP, strconv.Itoa(baseCfg.LocalPort))
	}
	psr.Plugin = baseCfg.Plugin.Type
	if s.Err == "" {
		remoteAddr := s.RemoteAddr
		// Prepend server address for TCP/UDP so the full endpoint is visible
		if serverAddr != "" && remoteAddr != "" && remoteAddr[0] == ':' {
			remoteAddr = serverAddr + remoteAddr
		}
		psr.RemoteAddr = remoteAddr
	}
	return psr
}

func handleProfileStatus(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		statuses := m.GetProfileProxyStatus(name)
		serverAddr := m.GetProfileServerAddr(name)
		if statuses == nil {
			statuses = []*proxy.WorkingStatus{}
		}
		res := make(model.StatusResp)
		for _, s := range statuses {
			res[s.Type] = append(res[s.Type], buildProxyStatusResp(s, serverAddr))
		}
		return res, nil
	}
}

// ─── Profile-scoped Proxy Handlers ─────────────────────

func handleListProfileProxies(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		proxies, err := m.ListProxies(name)
		if err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		resp := model.ProxyListResp{Proxies: make([]model.ProxyDefinition, 0, len(proxies))}
		for _, p := range proxies {
			payload, err := model.ProxyDefinitionFromConfigurer(p)
			if err != nil {
				return nil, httppkg.NewError(http.StatusInternalServerError, err.Error())
			}
			resp.Proxies = append(resp.Proxies, payload)
		}
		return resp, nil
	}
}

func handleCreateProfileProxy(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		body, err := ctx.Body()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "read body error: "+err.Error())
		}
		var payload model.ProxyDefinition
		if err := jsonx.Unmarshal(body, &payload); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "parse JSON error: "+err.Error())
		}
		if err := payload.Validate("", false); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		cfg, err := payload.ToConfigurer()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		runtimeCfg := cfg.Clone()
		runtimeCfg.Complete()
		if err := validation.ValidateProxyConfigurerForClient(runtimeCfg); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "validation error: "+err.Error())
		}
		if err := m.CreateProxy(profileName, cfg); err != nil {
			return nil, httppkg.NewError(http.StatusConflict, err.Error())
		}
		return payload, nil
	}
}

func handleGetProfileProxy(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		proxyName := ctx.Param("proxyName")
		p, err := m.GetProxy(profileName, proxyName)
		if err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		return model.ProxyDefinitionFromConfigurer(p)
	}
}

func handleUpdateProfileProxy(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		proxyName := ctx.Param("proxyName")
		body, err := ctx.Body()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "read body error: "+err.Error())
		}
		var payload model.ProxyDefinition
		if err := jsonx.Unmarshal(body, &payload); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "parse JSON error: "+err.Error())
		}
		if err := payload.Validate(proxyName, true); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		cfg, err := payload.ToConfigurer()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		runtimeCfg := cfg.Clone()
		runtimeCfg.Complete()
		if err := validation.ValidateProxyConfigurerForClient(runtimeCfg); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "validation error: "+err.Error())
		}
		if err := m.UpdateProxy(profileName, cfg); err != nil {
			return nil, httppkg.NewError(http.StatusInternalServerError, err.Error())
		}
		return payload, nil
	}
}

func handleDeleteProfileProxy(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		proxyName := ctx.Param("proxyName")
		if err := m.DeleteProxy(profileName, proxyName); err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		return nil, nil
	}
}

// ─── Profile-scoped Visitor Handlers ────────────────────

func handleListProfileVisitors(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		visitors, err := m.ListVisitors(name)
		if err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		resp := model.VisitorListResp{Visitors: make([]model.VisitorDefinition, 0, len(visitors))}
		for _, v := range visitors {
			payload, err := model.VisitorDefinitionFromConfigurer(v)
			if err != nil {
				return nil, httppkg.NewError(http.StatusInternalServerError, err.Error())
			}
			resp.Visitors = append(resp.Visitors, payload)
		}
		return resp, nil
	}
}

func handleCreateProfileVisitor(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		body, err := ctx.Body()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "read body error: "+err.Error())
		}
		var payload model.VisitorDefinition
		if err := jsonx.Unmarshal(body, &payload); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "parse JSON error: "+err.Error())
		}
		if err := payload.Validate("", false); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		cfg, err := payload.ToConfigurer()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		if err := m.CreateVisitor(profileName, cfg); err != nil {
			return nil, httppkg.NewError(http.StatusConflict, err.Error())
		}
		return payload, nil
	}
}

func handleGetProfileVisitor(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		visitorName := ctx.Param("visitorName")
		v, err := m.GetVisitor(profileName, visitorName)
		if err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		return model.VisitorDefinitionFromConfigurer(v)
	}
}

func handleUpdateProfileVisitor(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		visitorName := ctx.Param("visitorName")
		body, err := ctx.Body()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "read body error: "+err.Error())
		}
		var payload model.VisitorDefinition
		if err := jsonx.Unmarshal(body, &payload); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, "parse JSON error: "+err.Error())
		}
		if err := payload.Validate(visitorName, true); err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		cfg, err := payload.ToConfigurer()
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		if err := m.UpdateVisitor(profileName, cfg); err != nil {
			return nil, httppkg.NewError(http.StatusInternalServerError, err.Error())
		}
		return payload, nil
	}
}

func handleDeleteProfileVisitor(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		profileName := ctx.Param("name")
		visitorName := ctx.Param("visitorName")
		if err := m.DeleteVisitor(profileName, visitorName); err != nil {
			return nil, httppkg.NewError(http.StatusNotFound, err.Error())
		}
		return nil, nil
	}
}

// ─── Backward Compat Handlers ──────────────────────────

func handleGetConfig(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		// In manager mode, there's no single config file.
		// Return the profiles data as the "config" for backward compat.
		profiles := m.ListProfiles()
		return map[string]any{
			"mode":     "manager",
			"profiles": profiles,
		}, nil
	}
}

func handlePutConfig(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		return nil, httppkg.NewError(http.StatusBadRequest,
			"Config file editing is not available in manager mode. Use Profiles and Proxies APIs instead.")
	}
}

func handleReload(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		return map[string]any{"message": "Reload not needed in manager mode - changes are applied immediately."}, nil
	}
}

// ─── Frps Dashboard Query Handlers ─────────────────────

func handleGetProfileUsedPorts(m *FrpcManager) httppkg.APIHandler {
	return func(ctx *httppkg.Context) (any, error) {
		name := ctx.Param("name")
		usedPorts, err := m.GetProfileUsedPorts(name)
		if err != nil {
			return nil, httppkg.NewError(http.StatusBadRequest, err.Error())
		}
		return usedPorts, nil
	}
}
