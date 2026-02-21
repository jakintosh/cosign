package app

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type PaginationView struct {
	Page       int
	PageSize   int
	Total      int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	PrevPage   int
	NextPage   int
}

func NewPaginationView(
	page int,
	pageSize int,
	total int,
) PaginationView {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	pagination := PaginationView{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	if totalPages > 0 && page > 1 {
		pagination.HasPrev = true
		pagination.PrevPage = page - 1
	}

	if totalPages > 0 && page < totalPages {
		pagination.HasNext = true
		pagination.NextPage = page + 1
	}

	return pagination
}

func formatUnixTime(ts int64) string {
	if ts <= 0 {
		return "-"
	}

	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}

func campaignDetailPath(campaignID string) string {
	return "/campaigns/" + url.PathEscape(campaignID)
}

func campaignsPagePath(page int) string {
	if page < 1 {
		page = 1
	}

	return fmt.Sprintf("/campaigns?page=%d", page)
}

func signaturesPagePath(campaignID string, page int) string {
	if page < 1 {
		page = 1
	}

	return fmt.Sprintf(
		"/campaigns/%s/signatures?page=%d",
		url.PathEscape(campaignID),
		page,
	)
}

func campaignLocationsPath(campaignID string) string {
	return "/campaigns/" + url.PathEscape(campaignID) + "/locations"
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}
