package main

import (
	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	cli *openai.Client
}

func NewClient(openaitoken string) (*Client, error) {
	client := openai.NewClient(openaitoken)

	return &Client{
		cli: client,
	}, nil
}
