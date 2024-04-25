package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/shurcooL/graphql"
)

func main() {

	endpoint := "https://hub.snapshot.org/graphql"
	client := graphql.NewClient(endpoint, nil)
	queryCount := 0

	spaces := []string{"curve.eth", "kleros.eth", "comp-vote.eth", "qrobot.eth"}

	var queryVotes struct {
		Votes []struct {
			Voter   graphql.String
			Created graphql.Int
		} `graphql:"votes(first:1000 where:{space:$space created_lt:$created_lt created_gt:$created_gt} orderBy:\"created\", orderDirection:desc)"`
	}

	var queryProposals struct {
		Proposals []struct {
			Author  graphql.String
			Created graphql.Int
		} `graphql:"proposals(first:1000 where:{space:$space created_lt:$created_lt created_gt:$created_gt})"`
	}

	variables := map[string]any{
		"space":      graphql.String(spaces[0]),
		"created_lt": graphql.NewInt(1800000000), //upper bound, far future
		"created_gt": graphql.NewInt(0),          //lower bound, start of 2022
	}

	res := make(map[string]int)

	for i := 0; i < len(spaces); i++ {
		variables["space"] = graphql.String(spaces[i])
		variables["created_lt"] = graphql.NewInt(1800000000)
		err := client.Query(context.Background(), &queryProposals, variables)
		queryCount++
		if err != nil {
			log.Fatal(err)
		}
		spaceRes := make(map[string]int)
		for _, proposal := range queryProposals.Proposals {
			res[string(proposal.Author)] += 100
			spaceRes[string(proposal.Author)] += 100
		}
		for {
			err = client.Query(context.Background(), &queryVotes, variables)
			if err != nil {
				log.Fatal(err)
			}
			queryCount++
			if queryCount >= 60 {
				time.Sleep(time.Minute)
				queryCount = 0
			}
			for _, vote := range queryVotes.Votes {
				res[string(vote.Voter)] += 1
				spaceRes[string(vote.Voter)] += 1
			}
			if len(queryVotes.Votes) < 1000 {
				break
			}
			variables["created_lt"] = queryVotes.Votes[len(queryVotes.Votes)-1].Created - 1
		}
		fmt.Println(spaces[i] + ": " + strconv.Itoa(len(spaceRes)) + " users voted or made proposals")
	}

	for key, count := range res {
		if count < 3 {
			delete(res, key)
		}
	}
	fmt.Println(strconv.Itoa(len(res)) + " users voted more than twice or made proposals")
}
