package main

import (
  "log"
  "./crawler"
)

func main() {
  tabelog, err := crawler.NewTabelog()
  if err != nil {
    log.Fatal(err)
  }

  restaurantInfo, err := tabelog.Get("新宿", 1)
  for _, restaurant := range(restaurantInfo.Item) {
    println("Get: " + restaurant.RestaurantName)
    if err = tabelog.Save("restaurants", &restaurant); err != nil {
      log.Fatal(err)
    }

    reviewInfo, err := tabelog.GetReviews(restaurant.Rcd)
    if err != nil {
      log.Fatal(err)
    }

    for _, review := range(reviewInfo.Item) {
      println("- review: " + review.Title)
      if err = tabelog.Save("reviews", &review); err != nil {
        log.Fatal(err)
      }
    }
  }
}
