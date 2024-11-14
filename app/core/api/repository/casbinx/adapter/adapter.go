package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/predicate"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthpermissionrule"

	"github.com/pkg/errors"
)

const (
	DefaultTableName = "sca_auth_permission_rule"
	DefaultDatabase  = "schisandra_album_cloud"
)

type Adapter struct {
	client *ent.Client
	ctx    context.Context

	filtered bool
}

type Filter struct {
	Ptype []string
	V0    []string
	V1    []string
	V2    []string
	V3    []string
	V4    []string
	V5    []string
}

type Option func(a *Adapter) error

func open(driverName, dataSourceName string) (*ent.Client, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	var drv dialect.Driver
	if driverName == "pgx" {
		drv = entsql.OpenDB(dialect.Postgres, db)
	} else {
		drv = entsql.OpenDB(driverName, db)
	}
	return ent.NewClient(ent.Driver(drv)), nil
}

// NewAdapter returns an adapter by driver name and data source string.
func NewAdapter(driverName, dataSourceName string, options ...Option) (*Adapter, error) {
	client, err := open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	a := &Adapter{
		client: client,
		ctx:    context.Background(),
	}
	for _, option := range options {
		if err := option(a); err != nil {
			return nil, err
		}
	}
	if err := client.Schema.Create(a.ctx); err != nil {
		return nil, err
	}
	return a, nil
}

// NewAdapterWithClient create an adapter with client passed in.
// This method does not ensure the existence of database, user should create database manually.
func NewAdapterWithClient(client *ent.Client, options ...Option) (*Adapter, error) {
	a := &Adapter{
		client: client,
		ctx:    context.Background(),
	}
	for _, option := range options {
		if err := option(a); err != nil {
			return nil, err
		}
	}
	if err := client.Schema.Create(a.ctx); err != nil {
		return nil, err
	}
	return a, nil
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(model model.Model) error {
	policies, err := a.client.ScaAuthPermissionRule.Query().Order(ent.Asc("id")).All(a.ctx)
	if err != nil {
		return err
	}
	for _, policy := range policies {
		loadPolicyLine(policy, model)
	}
	return nil
}

// LoadFilteredPolicy loads only policy rules that match the filter.
// Filter parameter here is a Filter structure
func (a *Adapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {

	filterValue, ok := filter.(Filter)
	if !ok {
		return fmt.Errorf("invalid filter type: %v", reflect.TypeOf(filter))
	}

	session := a.client.ScaAuthPermissionRule.Query()
	if len(filterValue.Ptype) != 0 {
		session.Where(scaauthpermissionrule.PtypeIn(filterValue.Ptype...))
	}
	if len(filterValue.V0) != 0 {
		session.Where(scaauthpermissionrule.V0In(filterValue.V0...))
	}
	if len(filterValue.V1) != 0 {
		session.Where(scaauthpermissionrule.V1In(filterValue.V1...))
	}
	if len(filterValue.V2) != 0 {
		session.Where(scaauthpermissionrule.V2In(filterValue.V2...))
	}
	if len(filterValue.V3) != 0 {
		session.Where(scaauthpermissionrule.V3In(filterValue.V3...))
	}
	if len(filterValue.V4) != 0 {
		session.Where(scaauthpermissionrule.V4In(filterValue.V4...))
	}
	if len(filterValue.V5) != 0 {
		session.Where(scaauthpermissionrule.V5In(filterValue.V5...))
	}

	lines, err := session.All(a.ctx)
	if err != nil {
		return err
	}

	for _, line := range lines {
		loadPolicyLine(line, model)
	}
	a.filtered = true

	return nil
}

// IsFiltered returns true if the loaded policy has been filtered.
func (a *Adapter) IsFiltered() bool {
	return a.filtered
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	return a.WithTx(func(tx *ent.Tx) error {
		if _, err := tx.ScaAuthPermissionRule.Delete().Exec(a.ctx); err != nil {
			return err
		}
		lines := make([]*ent.ScaAuthPermissionRuleCreate, 0)

		for ptype, ast := range model["p"] {
			for _, policy := range ast.Policy {
				line := a.savePolicyLine(tx, ptype, policy)
				lines = append(lines, line)
			}
		}

		for ptype, ast := range model["g"] {
			for _, policy := range ast.Policy {
				line := a.savePolicyLine(tx, ptype, policy)
				lines = append(lines, line)
			}
		}

		_, err := tx.ScaAuthPermissionRule.CreateBulk(lines...).Save(a.ctx)
		return err
	})
}

// AddPolicy adds a policy rule to the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return a.WithTx(func(tx *ent.Tx) error {
		_, err := a.savePolicyLine(tx, ptype, rule).Save(a.ctx)
		return err
	})
}

// RemovePolicy removes a policy rule from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return a.WithTx(func(tx *ent.Tx) error {
		instance := a.toInstance(ptype, rule)
		_, err := tx.ScaAuthPermissionRule.Delete().Where(
			scaauthpermissionrule.PtypeEQ(instance.Ptype),
			scaauthpermissionrule.V0EQ(instance.V0),
			scaauthpermissionrule.V1EQ(instance.V1),
			scaauthpermissionrule.V2EQ(instance.V2),
			scaauthpermissionrule.V3EQ(instance.V3),
			scaauthpermissionrule.V4EQ(instance.V4),
			scaauthpermissionrule.V5EQ(instance.V5),
		).Exec(a.ctx)
		return err
	})
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return a.WithTx(func(tx *ent.Tx) error {
		cond := make([]predicate.ScaAuthPermissionRule, 0)
		cond = append(cond, scaauthpermissionrule.PtypeEQ(ptype))
		if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) && len(fieldValues[0-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V0EQ(fieldValues[0-fieldIndex]))
		}
		if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) && len(fieldValues[1-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V1EQ(fieldValues[1-fieldIndex]))
		}
		if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) && len(fieldValues[2-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V2EQ(fieldValues[2-fieldIndex]))
		}
		if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) && len(fieldValues[3-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V3EQ(fieldValues[3-fieldIndex]))
		}
		if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) && len(fieldValues[4-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V4EQ(fieldValues[4-fieldIndex]))
		}
		if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) && len(fieldValues[5-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V5EQ(fieldValues[5-fieldIndex]))
		}
		_, err := tx.ScaAuthPermissionRule.Delete().Where(
			cond...,
		).Exec(a.ctx)
		return err
	})
}

// AddPolicies adds policy rules to the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	return a.WithTx(func(tx *ent.Tx) error {
		return a.createPolicies(tx, ptype, rules)
	})
}

// RemovePolicies removes policy rules from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	return a.WithTx(func(tx *ent.Tx) error {
		for _, rule := range rules {
			instance := a.toInstance(ptype, rule)
			if _, err := tx.ScaAuthPermissionRule.Delete().Where(
				scaauthpermissionrule.PtypeEQ(instance.Ptype),
				scaauthpermissionrule.V0EQ(instance.V0),
				scaauthpermissionrule.V1EQ(instance.V1),
				scaauthpermissionrule.V2EQ(instance.V2),
				scaauthpermissionrule.V3EQ(instance.V3),
				scaauthpermissionrule.V4EQ(instance.V4),
				scaauthpermissionrule.V5EQ(instance.V5),
			).Exec(a.ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *Adapter) WithTx(fn func(tx *ent.Tx) error) error {
	tx, err := a.client.Tx(a.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = errors.Wrapf(err, "rolling back transaction: %v", rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "committing transaction: %v", err)
	}
	return nil
}

func loadPolicyLine(line *ent.ScaAuthPermissionRule, model model.Model) {
	var p = []string{line.Ptype,
		line.V0, line.V1, line.V2, line.V3, line.V4, line.V5}

	var lineText string
	if line.V5 != "" {
		lineText = strings.Join(p, ", ")
	} else if line.V4 != "" {
		lineText = strings.Join(p[:6], ", ")
	} else if line.V3 != "" {
		lineText = strings.Join(p[:5], ", ")
	} else if line.V2 != "" {
		lineText = strings.Join(p[:4], ", ")
	} else if line.V1 != "" {
		lineText = strings.Join(p[:3], ", ")
	} else if line.V0 != "" {
		lineText = strings.Join(p[:2], ", ")
	}

	persist.LoadPolicyLine(lineText, model)
}

func (a *Adapter) toInstance(ptype string, rule []string) *ent.ScaAuthPermissionRule {
	instance := &ent.ScaAuthPermissionRule{}

	instance.Ptype = ptype

	if len(rule) > 0 {
		instance.V0 = rule[0]
	}
	if len(rule) > 1 {
		instance.V1 = rule[1]
	}
	if len(rule) > 2 {
		instance.V2 = rule[2]
	}
	if len(rule) > 3 {
		instance.V3 = rule[3]
	}
	if len(rule) > 4 {
		instance.V4 = rule[4]
	}
	if len(rule) > 5 {
		instance.V5 = rule[5]
	}
	return instance
}

func (a *Adapter) savePolicyLine(tx *ent.Tx, ptype string, rule []string) *ent.ScaAuthPermissionRuleCreate {
	line := tx.ScaAuthPermissionRule.Create()

	line.SetPtype(ptype)
	if len(rule) > 0 {
		line.SetV0(rule[0])
	}
	if len(rule) > 1 {
		line.SetV1(rule[1])
	}
	if len(rule) > 2 {
		line.SetV2(rule[2])
	}
	if len(rule) > 3 {
		line.SetV3(rule[3])
	}
	if len(rule) > 4 {
		line.SetV4(rule[4])
	}
	if len(rule) > 5 {
		line.SetV5(rule[5])
	}

	return line
}

// UpdatePolicy updates a policy rule from storage.
// This is part of the Auto-Save feature.
func (a *Adapter) UpdatePolicy(sec string, ptype string, oldRule, newPolicy []string) error {
	return a.WithTx(func(tx *ent.Tx) error {
		rule := a.toInstance(ptype, oldRule)
		line := tx.ScaAuthPermissionRule.Update().Where(
			scaauthpermissionrule.PtypeEQ(rule.Ptype),
			scaauthpermissionrule.V0EQ(rule.V0),
			scaauthpermissionrule.V1EQ(rule.V1),
			scaauthpermissionrule.V2EQ(rule.V2),
			scaauthpermissionrule.V3EQ(rule.V3),
			scaauthpermissionrule.V4EQ(rule.V4),
			scaauthpermissionrule.V5EQ(rule.V5),
		)
		rule = a.toInstance(ptype, newPolicy)
		line.SetV0(rule.V0)
		line.SetV1(rule.V1)
		line.SetV2(rule.V2)
		line.SetV3(rule.V3)
		line.SetV4(rule.V4)
		line.SetV5(rule.V5)
		_, err := line.Save(a.ctx)
		return err
	})
}

// UpdatePolicies updates some policy rules to storage, like db, redis.
func (a *Adapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	return a.WithTx(func(tx *ent.Tx) error {
		for _, policy := range oldRules {
			rule := a.toInstance(ptype, policy)
			if _, err := tx.ScaAuthPermissionRule.Delete().Where(
				scaauthpermissionrule.PtypeEQ(rule.Ptype),
				scaauthpermissionrule.V0EQ(rule.V0),
				scaauthpermissionrule.V1EQ(rule.V1),
				scaauthpermissionrule.V2EQ(rule.V2),
				scaauthpermissionrule.V3EQ(rule.V3),
				scaauthpermissionrule.V4EQ(rule.V4),
				scaauthpermissionrule.V5EQ(rule.V5),
			).Exec(a.ctx); err != nil {
				return err
			}
		}
		lines := make([]*ent.ScaAuthPermissionRuleCreate, 0)
		for _, policy := range newRules {
			lines = append(lines, a.savePolicyLine(tx, ptype, policy))
		}
		if _, err := tx.ScaAuthPermissionRule.CreateBulk(lines...).Save(a.ctx); err != nil {
			return err
		}
		return nil
	})
}

// UpdateFilteredPolicies deletes old rules and adds new rules.
func (a *Adapter) UpdateFilteredPolicies(sec string, ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	oldPolicies := make([][]string, 0)
	err := a.WithTx(func(tx *ent.Tx) error {
		cond := make([]predicate.ScaAuthPermissionRule, 0)
		cond = append(cond, scaauthpermissionrule.PtypeEQ(ptype))
		if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) && len(fieldValues[0-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V0EQ(fieldValues[0-fieldIndex]))
		}
		if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) && len(fieldValues[1-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V1EQ(fieldValues[1-fieldIndex]))
		}
		if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) && len(fieldValues[2-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V2EQ(fieldValues[2-fieldIndex]))
		}
		if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) && len(fieldValues[3-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V3EQ(fieldValues[3-fieldIndex]))
		}
		if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) && len(fieldValues[4-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V4EQ(fieldValues[4-fieldIndex]))
		}
		if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) && len(fieldValues[5-fieldIndex]) > 0 {
			cond = append(cond, scaauthpermissionrule.V5EQ(fieldValues[5-fieldIndex]))
		}
		rules, err := tx.ScaAuthPermissionRule.Query().
			Where(cond...).
			All(a.ctx)
		if err != nil {
			return err
		}
		ruleIDs := make([]int, 0, len(rules))
		for _, r := range rules {
			ruleIDs = append(ruleIDs, r.ID)
		}

		_, err = tx.ScaAuthPermissionRule.Delete().
			Where(scaauthpermissionrule.IDIn(ruleIDs...)).
			Exec(a.ctx)
		if err != nil {
			return err
		}

		if err := a.createPolicies(tx, ptype, newPolicies); err != nil {
			return err
		}
		for _, rule := range rules {
			oldPolicies = append(oldPolicies, ScaAuthPermissionRuleToStringArray(rule))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return oldPolicies, nil
}

func (a *Adapter) createPolicies(tx *ent.Tx, ptype string, policies [][]string) error {
	lines := make([]*ent.ScaAuthPermissionRuleCreate, 0)
	for _, policy := range policies {
		lines = append(lines, a.savePolicyLine(tx, ptype, policy))
	}
	if _, err := tx.ScaAuthPermissionRule.CreateBulk(lines...).Save(a.ctx); err != nil {
		return err
	}
	return nil
}

func ScaAuthPermissionRuleToStringArray(rule *ent.ScaAuthPermissionRule) []string {
	arr := make([]string, 0)
	if rule.V0 != "" {
		arr = append(arr, rule.V0)
	}
	if rule.V1 != "" {
		arr = append(arr, rule.V1)
	}
	if rule.V2 != "" {
		arr = append(arr, rule.V2)
	}
	if rule.V3 != "" {
		arr = append(arr, rule.V3)
	}
	if rule.V4 != "" {
		arr = append(arr, rule.V4)
	}
	if rule.V5 != "" {
		arr = append(arr, rule.V5)
	}
	return arr
}
