package grpc

import (
	"context"
	"fmt"

	"github.com/konveyor/analyzer-lsp/provider"
	pb "github.com/konveyor/analyzer-lsp/provider/internal/grpc"
	"go.lsp.dev/uri"
)

type grpcServiceClient struct {
	id     int64
	ctx    context.Context
	config provider.InitConfig
	client pb.ProviderServiceClient
}

var _ provider.ServiceClient = &grpcServiceClient{}

func (g *grpcServiceClient) Evaluate(cap string, conditionInfo []byte) (provider.ProviderEvaluateResponse, error) {
	m := pb.EvaluateRequest{
		Cap:           cap,
		ConditionInfo: string(conditionInfo),
		Id:            g.id,
	}
	r, err := g.client.Evaluate(g.ctx, &m)
	if err != nil {
		return provider.ProviderEvaluateResponse{}, err
	}

	if !r.Successful {
		return provider.ProviderEvaluateResponse{}, fmt.Errorf(r.Error)
	}

	if !r.Response.Matched {
		return provider.ProviderEvaluateResponse{
			Matched:         false,
			TemplateContext: r.Response.TemplateContext.AsMap(),
		}, nil
	}

	incs := []provider.IncidentContext{}
	for _, i := range r.Response.IncidentContexts {
		inc := provider.IncidentContext{
			FileURI:   uri.URI(i.FileURI),
			Variables: i.GetVariables().AsMap(),
		}
		if i.Effort != nil {
			num := int(*i.Effort)
			inc.Effort = &num
		}
		links := []provider.ExternalLinks{}
		for _, l := range i.Links {
			links = append(links, provider.ExternalLinks{
				URL:   l.Url,
				Title: l.Title,
			})
		}
		inc.Links = links
		if i.CodeLocation != nil {
			inc.CodeLocation = &provider.Location{
				StartPosition: provider.Position{
					Line:      i.CodeLocation.StartPosition.Line,
					Character: i.CodeLocation.StartPosition.Character,
				},
				EndPosition: provider.Position{
					Line:      i.CodeLocation.EndPosition.Line,
					Character: i.CodeLocation.EndPosition.Character,
				},
			}
		}
		incs = append(incs, inc)
	}

	return provider.ProviderEvaluateResponse{
		Matched:         true,
		Incidents:       incs,
		TemplateContext: r.Response.TemplateContext.AsMap(),
	}, nil
}

// We don't have dependencies
func (g *grpcServiceClient) GetDependencies() ([]provider.Dep, uri.URI, error) {
	d, err := g.client.GetDependencies(g.ctx, &pb.ServiceRequest{Id: g.id})
	if err != nil {
		return nil, uri.URI(""), err
	}
	if !d.Successful {
		return nil, uri.URI(""), fmt.Errorf(d.Error)
	}

	provs := []provider.Dep{}
	for _, x := range d.List.Deps {
		provs = append(provs, provider.Dep{
			Name:               x.Name,
			Version:            x.Version,
			Type:               x.Type,
			Indirect:           x.Indirect,
			ResolvedIdentifier: x.ResolvedIdentifier,
			Extras:             x.Extras.AsMap(),
		})
	}

	u, err := uri.Parse(d.FileURI)
	if err != nil {
		u = uri.URI(d.FileURI)
	}

	return provs, u, nil

}

func recreateDAGAddedItems(items []*pb.DependencyDAGItem) []provider.DepDAGItem {

	deps := []provider.DepDAGItem{}
	for _, x := range items {
		deps = append(deps, provider.DepDAGItem{
			Dep: provider.Dep{
				Name:               x.Key.Name,
				Version:            x.Key.Version,
				Type:               x.Key.Type,
				Indirect:           x.Key.Indirect,
				ResolvedIdentifier: x.Key.ResolvedIdentifier,
				Extras:             x.Key.Extras.AsMap(),
			},
			AddedDeps: recreateDAGAddedItems(x.AddedDeps),
		})
	}
	return deps
}

// We don't have dependencies
func (g *grpcServiceClient) GetDependenciesDAG() ([]provider.DepDAGItem, uri.URI, error) {
	d, err := g.client.GetDependenciesDAG(g.ctx, &pb.ServiceRequest{Id: g.id})
	if err != nil {
		return nil, uri.URI(""), err
	}
	if !d.Successful {
		return nil, uri.URI(""), fmt.Errorf(d.Error)
	}
	m := recreateDAGAddedItems(d.List)

	u, err := uri.Parse(d.FileURI)
	if err != nil {
		return nil, uri.URI(""), fmt.Errorf(d.Error)
	}

	return m, u, nil

}

func (g *grpcServiceClient) Stop() {
	g.client.Stop(context.TODO(), &pb.ServiceRequest{Id: g.id})
}
